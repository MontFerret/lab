package sources_test

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/gobwas/glob"

	"github.com/MontFerret/lab/v2/pkg/sources"
)

func TestSourceFilters(t *testing.T) {
	names := []string{
		"root.pick.fql",
		"nested/child.pick.fql",
		"nested/deep/leaf.pick.fql",
		"nested/skip.fql",
		"suite.yaml",
		"suite.yml",
		"ignored.pick.txt",
	}
	root := newFilterDirectory(t, names)
	repo := newFilterRepository(t, names)

	for _, backend := range []string{"filesystem", "git"} {
		t.Run(backend, func(t *testing.T) {
			nestedPattern := "nested/*"
			if backend == "filesystem" {
				nestedPattern = glob.QuoteMeta(filepath.Join(root, "nested")) + "*"
			}

			for _, tc := range []struct {
				name    string
				pattern string
				want    []string
			}{
				{name: "no filter", want: names[:6]},
				{name: "wildcard crosses directories", pattern: "*.pick.fql", want: names[:3]},
				{name: "nested path", pattern: nestedPattern, want: names[1:4]},
				{name: "no matches", pattern: "*.missing.fql"},
			} {
				t.Run(tc.name, func(t *testing.T) {
					var src sources.Source
					if backend == "filesystem" {
						src = newFilteredFileSystem(t, root, tc.pattern)
					} else {
						src = newFilteredGit(t, repo, tc.pattern)
					}

					files := readFilterFiles(t, src)
					got := make([]string, 0, len(files))

					for _, file := range files {
						name := file.Name
						if backend == "filesystem" {
							relative, err := filepath.Rel(root, name)
							if err != nil {
								t.Fatal(err)
							}

							name = filepath.ToSlash(relative)
						}

						if file.Source != src || string(file.Content) != "RETURN true" {
							t.Errorf("source identity or content changed for %q", file.Name)
						}

						got = append(got, name)
					}

					slices.Sort(got)
					want := slices.Clone(tc.want)
					slices.Sort(want)

					if !slices.Equal(got, want) {
						t.Fatalf("files = %v, want %v", got, want)
					}
				})
			}
		})
	}
}

func TestFileSystemSingleFileFilter(t *testing.T) {
	root := newFilterDirectory(t, []string{"query.fql"})

	for _, tc := range []struct {
		query string
		want  int
	}{
		{query: "", want: 1},
		{query: "filter=", want: 1},
		{query: "filter=*.fql", want: 1},
		{query: "filter=*.yaml", want: 0},
	} {
		t.Run(tc.query, func(t *testing.T) {
			src, err := sources.NewFileSystem(&url.URL{Path: filepath.Join(root, "query.fql"), RawQuery: tc.query})
			if err != nil {
				t.Fatal(err)
			}

			files := readFilterFiles(t, src)
			if len(files) != tc.want {
				t.Fatalf("found %d files, want %d", len(files), tc.want)
			}
		})
	}
}

func TestSourceFilterRejectsMalformedPattern(t *testing.T) {
	for _, tc := range []struct {
		name string
		new  sources.SourceFactory
	}{
		{name: "filesystem", new: sources.NewFileSystem},
		{name: "git", new: sources.NewGit},
	} {
		t.Run(tc.name, func(t *testing.T) {
			src, err := tc.new(&url.URL{Path: "unused", RawQuery: url.Values{"filter": {"["}}.Encode()})
			if err == nil || src != nil {
				t.Fatalf("malformed filter returned source %v, error %v", src, err)
			}
		})
	}
}

func BenchmarkSourceFilteredRead(b *testing.B) {
	names := make([]string, 100)
	for i := range names {
		suffix := "skip"
		if i%2 == 0 {
			suffix = "pick"
		}

		names[i] = fmt.Sprintf("nested/query-%03d.%s.fql", i, suffix)
	}

	b.Run("filesystem", func(b *testing.B) {
		root := newFilterDirectory(b, names)
		src := newFilteredFileSystem(b, root, "*.pick.fql")
		b.ReportAllocs()

		for b.Loop() {
			if files := readFilterFiles(b, src); len(files) != 50 {
				b.Fatalf("found %d files, want 50", len(files))
			}
		}
	})

	b.Run("git", func(b *testing.B) {
		repo := newFilterRepository(b, names)
		src := newFilteredGit(b, repo, "*.pick.fql")
		b.ReportAllocs()

		for b.Loop() {
			if files := readFilterFiles(b, src); len(files) != 50 {
				b.Fatalf("found %d files, want 50", len(files))
			}
		}
	})
}

func newFilterDirectory(tb testing.TB, names []string) string {
	tb.Helper()
	root := tb.TempDir()

	for _, name := range names {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			tb.Fatal(err)
		}

		if err := os.WriteFile(path, []byte("RETURN true"), 0o644); err != nil {
			tb.Fatal(err)
		}
	}

	return root
}

func newFilterRepository(tb testing.TB, names []string) *git.Repository {
	tb.Helper()

	repo, err := git.Init(memory.NewStorage(), memfs.New())
	if err != nil {
		tb.Fatal(err)
	}

	tree, err := repo.Worktree()
	if err != nil {
		tb.Fatal(err)
	}

	for _, name := range names {
		file, err := tree.Filesystem.Create(name)
		if err != nil {
			tb.Fatal(err)
		}

		_, writeErr := file.Write([]byte("RETURN true"))
		closeErr := file.Close()
		if writeErr != nil || closeErr != nil {
			tb.Fatalf("write fixture: %v; close fixture: %v", writeErr, closeErr)
		}

		if _, err := tree.Add(name); err != nil {
			tb.Fatal(err)
		}
	}

	_, err = tree.Commit("Filter fixtures", &git.CommitOptions{
		Author: &object.Signature{Name: "Lab", Email: "lab@montferret.com", When: time.Unix(1, 0)},
	})
	if err != nil {
		tb.Fatal(err)
	}

	return repo
}

func newFilteredFileSystem(tb testing.TB, path, pattern string) sources.Source {
	tb.Helper()

	src, err := sources.NewFileSystem(&url.URL{Path: path, RawQuery: url.Values{"filter": {pattern}}.Encode()})
	if err != nil {
		tb.Fatal(err)
	}

	return src
}

func newFilteredGit(tb testing.TB, repo *git.Repository, pattern string) sources.Source {
	tb.Helper()

	if pattern == "" {
		src, err := sources.NewGitFrom(repo, nil)
		if err != nil {
			tb.Fatal(err)
		}

		return src
	}

	src, err := sources.NewGitFrom(repo, glob.MustCompile(pattern))
	if err != nil {
		tb.Fatal(err)
	}

	return src
}

func readFilterFiles(tb testing.TB, src sources.Source) []sources.File {
	tb.Helper()
	files, errs := src.Read(tb.Context())
	var result []sources.File

	for files != nil || errs != nil {
		select {
		case file, ok := <-files:
			if !ok {
				files = nil
				continue
			}

			result = append(result, file)
		case err, ok := <-errs:
			if !ok {
				errs = nil
				continue
			}

			tb.Error(err)
		case <-tb.Context().Done():
			tb.Fatal(tb.Context().Err())
		}
	}

	return result
}
