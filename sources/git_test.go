package sources_test

import (
	"context"
	"fmt"
	"github.com/gobwas/glob"
	"net/url"
	"testing"
	"time"

	"github.com/MontFerret/lab/sources"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
	. "github.com/smartystreets/goconvey/convey"
)

type MockFile struct {
	Name    string
	Content string
}

func TestGit(t *testing.T) {
	initRepo := func(files []MockFile) (*git.Repository, error) {
		repo, err := git.Init(memory.NewStorage(), memfs.New())

		if err != nil {
			return nil, err
		}

		tree, err := repo.Worktree()

		if err != nil {
			return nil, err
		}

		for i, f := range files {
			file, err := tree.Filesystem.Create(f.Name)

			if err != nil {
				return nil, err
			}

			content := f.Content

			if content == "" {
				content = fmt.Sprintf("RETURN %d", i)
			}

			_, err = file.Write([]byte(content))

			if err != nil {
				return nil, err
			}

			if err := file.Close(); err != nil {
				return nil, err
			}
		}

		if err := tree.AddGlob("*.*"); err != nil {
			return nil, err
		}

		_, err = tree.Commit("Initial commit", &git.CommitOptions{
			All: true,
			Author: &object.Signature{
				Name:  "Lab",
				Email: "lab@montferret.com",
				When:  time.Now(),
			},
			Committer: &object.Signature{
				Name:  "Lab",
				Email: "lab@montferret.com",
				When:  time.Now(),
			},
		})

		if err != nil {
			return nil, err
		}

		return repo, nil
	}

	Convey("Git source", t, func() {
		Convey(".Read", func() {
			Convey("When a repo does not exist", func() {
				Convey("Should send an error", func() {
					u, err := url.Parse("http://localhost/user/repo")
					So(err, ShouldBeNil)
					src, err := sources.NewGit(*u)
					So(err, ShouldBeNil)

					onNext, onError := src.Read(context.Background())

					select {
					case e := <-onError:
						So(e, ShouldNotBeNil)
					case <-onNext:
						panic("Should not return files")
					}
				})
			})

			Convey("Should read content if .fql", func() {
				repo, err := initRepo([]MockFile{
					{
						Name:    "query-1.fql",
						Content: "RETURN 'file1'",
					},
				})

				So(err, ShouldBeNil)

				src, err := sources.NewGitFrom(repo, nil)
				So(err, ShouldBeNil)

				onNext, onError := src.Read(context.Background())

				So(onNext, ShouldNotBeNil)
				So(onError, ShouldNotBeNil)

				select {
				case e := <-onError:
					panic(e)
				case f := <-onNext:
					So(string(f.Content), ShouldNotBeEmpty)
				}
			})

			Convey("Should ignore non-supported files", func() {
				repo, err := initRepo([]MockFile{
					{
						Name: "query-1.fql",
					},
					{
						Name: "query-2.aql",
					},
					{
						Name: "suite-1.yaml",
					},
					{
						Name: "suite-2.yml",
					},
				})

				So(err, ShouldBeNil)

				src, err := sources.NewGitFrom(repo, nil)
				So(err, ShouldBeNil)

				onNext, onError := src.Read(context.Background())

				So(onNext, ShouldNotBeNil)
				So(onError, ShouldNotBeNil)

				files := make([]string, 0, 4)

				var done bool

				for !done {
					select {
					case e, open := <-onError:
						if !open {
							break
						}

						panic(e)
					case f, open := <-onNext:
						if !open {
							done = true
							break
						}

						files = append(files, f.Name)
					}
				}

				So(files, ShouldHaveLength, 3)
			})

			Convey("When a ?filter defined", func() {
				Convey("Should filter out files", func() {
					repo, err := initRepo([]MockFile{
						{
							Name: "query-1.fql",
						},
						{
							Name: "query-2.pick.fql",
						},
						{
							Name: "query-3.fql",
						},
						{
							Name: "query-4.pick.fql",
						},
						{
							Name: "query-5.fql",
						},
					})

					So(err, ShouldBeNil)

					src, err := sources.NewGitFrom(repo, glob.MustCompile("*.pick.fql"))
					So(err, ShouldBeNil)

					onNext, onError := src.Read(context.Background())

					So(onNext, ShouldNotBeNil)
					So(onError, ShouldNotBeNil)

					files := make([]string, 0, 2)

					var done bool

					for !done {
						select {
						case e, open := <-onError:
							if !open {
								break
							}

							panic(e)
						case f, open := <-onNext:
							if !open {
								done = true
								break
							}

							files = append(files, f.Name)
						}
					}

					So(files, ShouldHaveLength, 2)
				})
			})
		})

		Convey(".Resolve", func() {
			repo, err := initRepo([]MockFile{
				{
					Name: "query-1.fql",
				},
				{
					Name: "query-2.fql",
				},
				{
					Name: "suite-1.yaml",
				},
				{
					Name: "suite-2.yml",
				},
			})

			So(err, ShouldBeNil)

			src, err := sources.NewGitFrom(repo, glob.MustCompile("*.pick.fql"))
			So(err, ShouldBeNil)

			onNext, onError := src.Resolve(context.Background(), "query-2.fql")

			So(onNext, ShouldNotBeNil)
			So(onError, ShouldNotBeNil)

			select {
			case e := <-onError:
				panic(e)
			case f := <-onNext:
				So(f.Name, ShouldEqual, "query-2.fql")
			}
		})
	})
}
