package staticserver_test

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/MontFerret/lab/v2/pkg/staticserver"
)

func TestServeEntries(t *testing.T) {
	Convey("Serve entry parsing", t, func() {
		root := t.TempDir()
		appDir := filepath.Join(root, "app")
		apiDir := filepath.Join(root, "api-mocks")
		filePath := filepath.Join(root, "hello.txt")

		So(os.Mkdir(appDir, 0o755), ShouldBeNil)
		So(os.Mkdir(apiDir, 0o755), ShouldBeNil)
		So(os.WriteFile(filePath, []byte("hello"), 0o644), ShouldBeNil)

		Convey("parses a path without alias or port", func() {
			entry, err := staticserver.ParseServeEntry(appDir)

			So(err, ShouldBeNil)
			So(entry.Path, ShouldEqual, appDir)
			So(entry.Alias, ShouldEqual, "app")
			So(entry.Port, ShouldEqual, 0)
		})

		Convey("parses a path with port", func() {
			entry, err := staticserver.ParseServeEntry(appDir + ":9090")

			So(err, ShouldBeNil)
			So(entry.Path, ShouldEqual, appDir)
			So(entry.Alias, ShouldEqual, "app")
			So(entry.Port, ShouldEqual, 9090)
		})

		Convey("parses a path with alias", func() {
			entry, err := staticserver.ParseServeEntry(appDir + "@frontend")

			So(err, ShouldBeNil)
			So(entry.Path, ShouldEqual, appDir)
			So(entry.Alias, ShouldEqual, "frontend")
			So(entry.Port, ShouldEqual, 0)
		})

		Convey("parses a path with alias and port", func() {
			entry, err := staticserver.ParseServeEntry(appDir + "@frontend:9090")

			So(err, ShouldBeNil)
			So(entry.Path, ShouldEqual, appDir)
			So(entry.Alias, ShouldEqual, "frontend")
			So(entry.Port, ShouldEqual, 9090)
		})

		Convey("keeps hyphenated aliases", func() {
			entry, err := staticserver.ParseServeEntry(appDir + "@api-mocks")

			So(err, ShouldBeNil)
			So(entry.Alias, ShouldEqual, "api-mocks")
		})

		Convey("derives the alias from a cleaned basename", func() {
			entry, err := staticserver.ParseServeEntry(apiDir + string(os.PathSeparator))

			So(err, ShouldBeNil)
			So(entry.Alias, ShouldEqual, "api-mocks")
		})

		Convey("rejects an invalid alias", func() {
			_, err := staticserver.ParseServeEntry(appDir + "@1app")

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, `invalid serve alias "1app"`)
		})

		Convey("rejects duplicate aliases", func() {
			_, err := staticserver.ParseServeEntries([]string{
				appDir + "@app",
				apiDir + "@app",
			})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, `duplicate static alias "app"`)
		})

		Convey("rejects a missing directory", func() {
			_, err := staticserver.ParseServeEntry(filepath.Join(root, "missing"))

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, `does not exist`)
		})

		Convey("rejects a file path", func() {
			_, err := staticserver.ParseServeEntry(filePath)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, `served directory "`+filePath+`" is not a directory`)
		})

		Convey("rejects an invalid port", func() {
			_, err := staticserver.ParseServeEntry(appDir + ":70000")

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, `invalid serve port "70000"`)
		})

		Convey("rejects legacy port-before-alias syntax", func() {
			_, err := staticserver.ParseServeEntry(appDir + ":9090@test")

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, `use <path>@<alias>:<port>`)
		})
	})
}
