package sources_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/MontFerret/lab/sources"
	. "github.com/smartystreets/goconvey/convey"
)

func TestFileSystem(t *testing.T) {
	Convey("Files system source", t, func() {
		Convey(".Read", func() {
			Convey("When a file does not exist", func() {
				Convey("Should send an error", func() {
					str := fmt.Sprintf("file://%stest.foo", os.TempDir())
					u, _ := url.Parse(str)
					src, err := sources.NewFileSystem(*u)

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
					file, err := ioutil.TempFile("", "lab.*.fql")

					So(err, ShouldBeNil)

					_, err = file.WriteString("RETURN 'foo'")
					file.Close()
					So(err, ShouldBeNil)

					defer os.Remove(file.Name())

					u, _ := url.Parse(fmt.Sprintf("file://%s", file.Name()))
					src, err := sources.NewFileSystem(*u)

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
					}
				})

				Convey("Should return error if not .fql", func() {
					file, err := ioutil.TempFile("", "lab.*.aql")

					So(err, ShouldBeNil)

					_, err = file.WriteString("RETURN 'foo'")
					file.Close()
					So(err, ShouldBeNil)

					defer os.Remove(file.Name())

					u, _ := url.Parse(fmt.Sprintf("file://%s", file.Name()))
					src, err := sources.NewFileSystem(*u)

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
						f, err := ioutil.TempFile("", "lab.*.fql")

						if err != nil {
							panic(err)
						}

						_, err = f.WriteString(fmt.Sprintf("RETURN %d", i))

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
						src, err := sources.NewFileSystem(*u)

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
						}
					}
				})
			})

			Convey("When a ?filter defined", func() {
				Convey("Should filter out files", func() {
					files := make([]string, 5)

					dir, err := ioutil.TempDir("", "lab_test")

					if err != nil {
						panic(err)
					}

					for i := range files {
						name := "lab.*.fql"

						if (i % 2) == 0 {
							name = "lab.*.take.fql"
						}

						f, err := ioutil.TempFile(dir, name)

						if err != nil {
							panic(err)
						}

						_, err = f.WriteString(fmt.Sprintf("RETURN %d", i))

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
					src, err := sources.NewFileSystem(*u)

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
					f1, err := ioutil.TempFile("", "lab.*.fql")
					So(err, ShouldBeNil)

					_, err = f1.WriteString("RETURN 'file1'")
					So(err, ShouldBeNil)
					So(f1.Close(), ShouldBeNil)

					f2, err := ioutil.TempFile("", "lab.*.fql")
					So(err, ShouldBeNil)

					_, err = f2.WriteString("RETURN 'file2'")
					So(err, ShouldBeNil)
					So(f2.Close(), ShouldBeNil)

					defer func() {
						os.Remove(f1.Name())
						os.Remove(f2.Name())
					}()

					u, _ := url.Parse(fmt.Sprintf("file://%s", f1.Name()))
					src, err := sources.NewFileSystem(*u)

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
					}
				})

				Convey("Should resolve a file from a different folder", func() {
					dir, err := ioutil.TempDir("", "lab-tests")
					So(err, ShouldBeNil)

					f1, err := ioutil.TempFile("", "lab.*.fql")
					So(err, ShouldBeNil)

					_, err = f1.WriteString("RETURN 'file1'")
					So(err, ShouldBeNil)
					So(f1.Close(), ShouldBeNil)

					f2, err := ioutil.TempFile(dir, "lab.*.fql")
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
					src, err := sources.NewFileSystem(*u)

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

					onNext, onError = src.Resolve(
						context.Background(),
						mustParseUrl(filepath.Join("./", filepath.Base(dir), filepath.Base(f2.Name()))),
					)

					select {
					case e := <-onError:
						So(e, ShouldBeNil)
					case f := <-onNext:
						So(string(f.Content), ShouldEqual, "RETURN 'file2'")
						So(f.Name, ShouldNotBeNil)
					}
				})
			})
		})
	})
}
