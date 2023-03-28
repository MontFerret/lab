package sources_test

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/MontFerret/lab/runtime"
	"github.com/MontFerret/lab/sources"
	T "github.com/MontFerret/lab/testing"
)

func TestFileSystem(t *testing.T) {
	Convey("Files system source", t, func() {
		Convey(".Read", func() {
			Convey("When a file does not exist", func() {
				Convey("Should send an error", func() {
					str := fmt.Sprintf("file://%stest.foo", os.TempDir())
					u, _ := url.Parse(str)
					src, err := sources.NewFileSystem(u)

					So(err, ShouldBeNil)
					So(src, ShouldNotBeNil)

					onNext, onError := src.Read(context.Background())

					So(onNext, ShouldNotBeNil)
					So(onError, ShouldNotBeNil)

					select {
					case e := <-onError:
						So(e, ShouldNotBeNil)
					case f := <-onNext:
						So(string(f.Content), ShouldBeEmpty)
					}
				})
			})

			Convey("When a single file passed", func() {
				Convey("Should read content if .fql", func() {
					file, err := os.CreateTemp("", "lab.*.fql")

					So(err, ShouldBeNil)

					_, err = file.WriteString("RETURN 'foo'")
					file.Close()
					So(err, ShouldBeNil)

					defer os.Remove(file.Name())

					u, _ := url.Parse(fmt.Sprintf("file://%s", file.Name()))
					src, err := sources.NewFileSystem(u)

					So(err, ShouldBeNil)
					So(src, ShouldNotBeNil)

					onNext, onError := src.Read(context.Background())

					So(onNext, ShouldNotBeNil)
					So(onError, ShouldNotBeNil)

					select {
					case e := <-onError:
						So(e, ShouldBeNil)
					case f := <-onNext:
						So(string(f.Content), ShouldEqual, "RETURN 'foo'")
						So(f.Name, ShouldNotBeNil)
						So(f.Source, ShouldEqual, src)
					}
				})

				Convey("Should return error if not .fql", func() {
					file, err := os.CreateTemp("", "lab.*.aql")

					So(err, ShouldBeNil)

					_, err = file.WriteString("RETURN 'foo'")
					file.Close()
					So(err, ShouldBeNil)

					defer os.Remove(file.Name())

					u, _ := url.Parse(fmt.Sprintf("file://%s", file.Name()))
					src, err := sources.NewFileSystem(u)

					So(err, ShouldBeNil)
					So(src, ShouldNotBeNil)

					onNext, onError := src.Read(context.Background())

					So(onNext, ShouldNotBeNil)
					So(onError, ShouldNotBeNil)

					select {
					case e := <-onError:
						So(e, ShouldNotBeNil)
					case f := <-onNext:
						So(string(f.Content), ShouldBeEmpty)
						So(f.Name, ShouldBeNil)
					}
				})
			})

			Convey("When a directory passed", func() {
				Convey("Should read all files", func() {
					files := make([]string, 5)

					for i := range files {
						f, err := os.CreateTemp("", "lab.*.fql")

						if err != nil {
							panic(err)
						}

						f.WriteString(fmt.Sprintf("RETURN %d", i))

						files[i] = f.Name()
						f.Close()
					}

					defer func() {
						for _, f := range files {
							os.Remove(f)
						}
					}()

					for i, f := range files {
						u, _ := url.Parse(fmt.Sprintf("file://%s", f))
						src, err := sources.NewFileSystem(u)

						So(err, ShouldBeNil)
						So(src, ShouldNotBeNil)

						onNext, onError := src.Read(context.Background())

						So(onNext, ShouldNotBeNil)
						So(onError, ShouldNotBeNil)

						select {
						case e := <-onError:
							So(e, ShouldBeNil)
						case f := <-onNext:
							So(string(f.Content), ShouldEqual, fmt.Sprintf("RETURN %d", i))
							So(f.Name, ShouldNotBeNil)
							So(f.Source, ShouldEqual, src)
						}
					}
				})
			})

			Convey("When a ?filter defined", func() {
				Convey("Should filter out files", func() {
					files := make([]string, 5)

					dir, err := os.MkdirTemp("", "lab_test")

					if err != nil {
						panic(err)
					}

					for i := range files {
						name := "lab.*.fql"

						if (i % 2) == 0 {
							name = "lab.*.take.fql"
						}

						f, err := os.CreateTemp(dir, name)

						if err != nil {
							panic(err)
						}

						f.WriteString(fmt.Sprintf("RETURN %d", i))

						files[i] = f.Name()
						f.Close()
					}

					defer func() {
						for _, f := range files {
							os.Remove(f)
						}

						os.Remove(dir)
					}()

					u, _ := url.Parse(fmt.Sprintf("file://%s?filter=**/*.take.fql", dir))
					src, err := sources.NewFileSystem(u)

					So(err, ShouldBeNil)
					So(src, ShouldNotBeNil)

					onNext, onError := src.Read(context.Background())

					So(onNext, ShouldNotBeNil)
					So(onError, ShouldNotBeNil)

					foundFiles := make([]sources.File, 0, len(files))

					var done bool

					for !done {
						select {
						case err, open := <-onError:
							if !open {
								done = true
								break
							}

							So(err, ShouldBeNil)
						case file, open := <-onNext:
							if !open {
								done = true
								break
							}

							So(file, ShouldNotBeNil)
							So(file.Source, ShouldEqual, src)

							foundFiles = append(foundFiles, file)
						}
					}

					So(foundFiles, ShouldHaveLength, 3)
				})
			})
		})

		Convey(".Resolve", func() {
			Convey("When a local file passed", func() {
				Convey("Should resolve a file from the same folder", func() {
					f1, err := os.CreateTemp("", "lab.*.fql")
					So(err, ShouldBeNil)

					_, err = f1.WriteString("RETURN 'file1'")
					So(err, ShouldBeNil)
					So(f1.Close(), ShouldBeNil)

					f2, err := os.CreateTemp("", "lab.*.fql")
					So(err, ShouldBeNil)

					_, err = f2.WriteString("RETURN 'file2'")
					So(err, ShouldBeNil)
					So(f2.Close(), ShouldBeNil)

					defer func() {
						os.Remove(f1.Name())
						os.Remove(f2.Name())
					}()

					u, _ := url.Parse(fmt.Sprintf("file://%s", f1.Name()))
					src, err := sources.NewFileSystem(u)

					So(err, ShouldBeNil)
					So(src, ShouldNotBeNil)

					onNext, onError := src.Read(context.Background())

					So(onNext, ShouldNotBeNil)
					So(onError, ShouldNotBeNil)

					select {
					case e := <-onError:
						So(e, ShouldBeNil)
					case f := <-onNext:
						So(string(f.Content), ShouldEqual, "RETURN 'file1'")
						So(f.Name, ShouldNotBeNil)
					}

					onNext, onError = src.Resolve(context.Background(), mustParseUrl(filepath.Base(f2.Name())))

					select {
					case e := <-onError:
						So(e, ShouldBeNil)
					case f := <-onNext:
						So(string(f.Content), ShouldEqual, "RETURN 'file2'")
						So(f.Name, ShouldNotBeNil)
						So(f.Source, ShouldEqual, src)
					}
				})

				Convey("Should resolve a file from a different folder", func() {
					dir, err := os.MkdirTemp("", "lab-tests")
					So(err, ShouldBeNil)

					f1, err := os.CreateTemp("", "lab.*.fql")
					So(err, ShouldBeNil)

					_, err = f1.WriteString("RETURN 'file1'")
					So(err, ShouldBeNil)
					So(f1.Close(), ShouldBeNil)

					f2, err := os.CreateTemp(dir, "lab.*.fql")
					So(err, ShouldBeNil)

					_, err = f2.WriteString("RETURN 'file2'")
					So(err, ShouldBeNil)
					So(f2.Close(), ShouldBeNil)

					defer func() {
						os.Remove(f1.Name())
						os.Remove(f2.Name())
						os.Remove(dir)
					}()

					u, _ := url.Parse(fmt.Sprintf("file://%s", f1.Name()))
					src, err := sources.NewFileSystem(u)

					So(err, ShouldBeNil)
					So(src, ShouldNotBeNil)

					onNext, onError := src.Read(context.Background())

					So(onNext, ShouldNotBeNil)
					So(onError, ShouldNotBeNil)

					select {
					case e := <-onError:
						So(e, ShouldBeNil)
					case f := <-onNext:
						So(string(f.Content), ShouldEqual, "RETURN 'file1'")
						So(f.Name, ShouldNotBeNil)
						So(f.Source, ShouldEqual, src)
					}

					path, err := filepath.Rel(filepath.Dir(f1.Name()), f2.Name())
					So(err, ShouldBeNil)

					onNext, onError = src.Resolve(
						context.Background(),
						mustParseUrl(path),
					)

					select {
					case e := <-onError:
						So(e, ShouldBeNil)
					case f := <-onNext:
						So(string(f.Content), ShouldEqual, "RETURN 'file2'")
						So(f.Name, ShouldNotBeNil)
						So(f.Source, ShouldEqual, src)
					}
				})

				Convey("Should resolve a file from a sibling folder", func() {
					dir1, err := os.MkdirTemp("", "lab-tests-*")
					So(err, ShouldBeNil)

					dir2, err := os.MkdirTemp("", "lab-tests-*")
					So(err, ShouldBeNil)

					f1, err := os.CreateTemp(dir1, "lab.*.fql")
					So(err, ShouldBeNil)

					_, err = f1.WriteString("RETURN 'file1'")
					So(err, ShouldBeNil)
					So(f1.Close(), ShouldBeNil)

					f2, err := os.CreateTemp(dir2, "lab.*.fql")
					So(err, ShouldBeNil)

					_, err = f2.WriteString("RETURN 'file2'")
					So(err, ShouldBeNil)
					So(f2.Close(), ShouldBeNil)

					defer func() {
						os.Remove(f1.Name())
						os.Remove(f2.Name())
						os.Remove(dir1)
						os.Remove(dir2)
					}()

					u, _ := url.Parse(fmt.Sprintf("file://%s", f1.Name()))
					src, err := sources.NewFileSystem(u)

					So(err, ShouldBeNil)
					So(src, ShouldNotBeNil)

					onNext, onError := src.Read(context.Background())

					So(onNext, ShouldNotBeNil)
					So(onError, ShouldNotBeNil)

					select {
					case e := <-onError:
						So(e, ShouldBeNil)
					case f := <-onNext:
						So(string(f.Content), ShouldEqual, "RETURN 'file1'")
						So(f.Name, ShouldNotBeNil)
						So(f.Source, ShouldEqual, src)
					}

					path, err := filepath.Rel(dir1, f2.Name())
					So(err, ShouldBeNil)
					onNext, onError = src.Resolve(
						context.Background(),
						mustParseUrl(path),
					)

					select {
					case e := <-onError:
						So(e, ShouldBeNil)
					case f := <-onNext:
						So(string(f.Content), ShouldEqual, "RETURN 'file2'")
						So(f.Name, ShouldNotBeNil)
						So(f.Source, ShouldEqual, src)
					}
				})

				Convey("Should resolve a file using only rel paths", func() {
					dir1, err := os.MkdirTemp("", "lab-tests-*")
					So(err, ShouldBeNil)

					dir2, err := os.MkdirTemp("", "lab-tests-*")
					So(err, ShouldBeNil)

					f1, err := os.CreateTemp(dir1, "lab.*.fql")
					So(err, ShouldBeNil)

					_, err = f1.WriteString("RETURN 'file1'")
					So(err, ShouldBeNil)
					So(f1.Close(), ShouldBeNil)

					f2, err := os.CreateTemp(dir2, "lab.*.fql")
					So(err, ShouldBeNil)

					_, err = f2.WriteString("RETURN 'file2'")
					So(err, ShouldBeNil)
					So(f2.Close(), ShouldBeNil)

					defer func() {
						os.Remove(f1.Name())
						os.Remove(f2.Name())
						os.Remove(dir1)
						os.Remove(dir2)
					}()

					wd, err := os.Getwd()
					So(err, ShouldBeNil)

					relToFile, err := filepath.Rel(wd, f1.Name())
					So(err, ShouldBeNil)

					u, _ := url.Parse(fmt.Sprintf("file://%s", relToFile))
					src, err := sources.NewFileSystem(u)

					So(err, ShouldBeNil)
					So(src, ShouldNotBeNil)

					onNext, onError := src.Read(context.Background())

					So(onNext, ShouldNotBeNil)
					So(onError, ShouldNotBeNil)

					select {
					case e := <-onError:
						So(e, ShouldBeNil)
					case f := <-onNext:
						So(string(f.Content), ShouldEqual, "RETURN 'file1'")
						So(f.Name, ShouldNotBeNil)
						So(f.Source, ShouldEqual, src)
					}

					path, err := filepath.Rel(dir1, f2.Name())
					So(err, ShouldBeNil)
					onNext, onError = src.Resolve(
						context.Background(),
						mustParseUrl(path),
					)

					select {
					case e := <-onError:
						So(e, ShouldBeNil)
					case f := <-onNext:
						So(string(f.Content), ShouldEqual, "RETURN 'file2'")
						So(f.Name, ShouldNotBeNil)
						So(f.Source, ShouldEqual, src)
					}
				})

				Convey("Should resolve from a path different from a based one", func() {
					dirTestsRoot, err := os.MkdirTemp("", "lab-tests-test-root-*")
					So(err, ShouldBeNil)

					dirTestsNested, err := os.MkdirTemp(dirTestsRoot, "lab-tests-test-nested-*")
					So(err, ShouldBeNil)

					dirTestsNested2x, err := os.MkdirTemp(dirTestsNested, "lab-tests-test-nested-2x-*")
					So(err, ShouldBeNil)

					dirExmRoot, err := os.MkdirTemp("", "lab-tests-examples-root-*")
					So(err, ShouldBeNil)

					f1, err := os.CreateTemp(dirExmRoot, "lab.*.fql")
					So(err, ShouldBeNil)

					_, err = f1.WriteString("RETURN 'file1'")
					So(err, ShouldBeNil)
					So(f1.Close(), ShouldBeNil)

					f2, err := os.CreateTemp(dirTestsNested2x, "lab.*.yaml")
					So(err, ShouldBeNil)

					f2Content := fmt.Sprintf(`
query:
  ref: ../../../%s/%s
assert:
  text: RETURN TRUE
`, filepath.Base(dirExmRoot), filepath.Base(f1.Name()))
					_, err = f2.WriteString(f2Content)
					So(err, ShouldBeNil)
					So(f2.Close(), ShouldBeNil)

					defer func() {
						os.Remove(f1.Name())
						os.Remove(f2.Name())
						os.Remove(dirTestsRoot)
						os.Remove(dirTestsNested)
						os.Remove(dirTestsNested2x)
						os.Remove(dirExmRoot)
					}()

					u, _ := url.Parse(fmt.Sprintf("file://%s", dirTestsRoot))
					src, err := sources.NewFileSystem(u)

					So(err, ShouldBeNil)
					So(src, ShouldNotBeNil)

					onNext, onError := src.Read(context.Background())

					So(onNext, ShouldNotBeNil)
					So(onError, ShouldNotBeNil)

					rt := runtime.AsFunc(func(ctx context.Context, query string, params map[string]interface{}) ([]byte, error) {
						return []byte(""), nil
					})

					for f := range onNext {
						s, err := T.NewSuite(T.Options{
							File:    f,
							Timeout: 0,
						})

						So(err, ShouldBeNil)

						err = s.Run(context.Background(), rt, T.NewParams())
						So(err, ShouldBeNil)
					}
				})

			})
		})
	})
}
