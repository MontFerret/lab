package cdn_test

import (
	"github.com/MontFerret/lab/cdn"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestDirectory(t *testing.T) {
	Convey("Directory", t, func() {
		Convey("NewDirectoryFrom", func() {
			Convey("Valid input", func() {
				Convey("When 'content'", func() {
					dir, err := cdn.NewDirectoryFrom("content")

					So(err, ShouldBeNil)
					So(dir.Path, ShouldEqual, "content")
					So(dir.Name, ShouldEqual, "content")
					So(dir.Port, ShouldBeGreaterThan, 0)
				})

				Convey("When 'lab/content'", func() {
					dir, err := cdn.NewDirectoryFrom("lab/content")

					So(err, ShouldBeNil)
					So(dir.Path, ShouldEqual, "lab/content")
					So(dir.Name, ShouldEqual, "content")
					So(dir.Port, ShouldBeGreaterThan, 0)
				})

				Convey("When 'lab/content:9090'", func() {
					dir, err := cdn.NewDirectoryFrom("lab/content:9090")

					So(err, ShouldBeNil)
					So(dir.Path, ShouldEqual, "lab/content")
					So(dir.Name, ShouldEqual, "content")
					So(dir.Port, ShouldEqual, 9090)
				})

				Convey("When 'lab/content:9090@test'", func() {
					dir, err := cdn.NewDirectoryFrom("lab/content:9090@test")

					So(err, ShouldBeNil)
					So(dir.Path, ShouldEqual, "lab/content")
					So(dir.Name, ShouldEqual, "test")
					So(dir.Port, ShouldEqual, 9090)
				})

				Convey("When 'lab/content@test:9090'", func() {
					dir, err := cdn.NewDirectoryFrom("lab/content@test:9090")

					So(err, ShouldBeNil)
					So(dir.Path, ShouldEqual, "lab/content")
					So(dir.Name, ShouldEqual, "test")
					So(dir.Port, ShouldBeGreaterThan, 0)
				})
			})
			Convey("Invalid input", func() {
				Convey("When ''", func() {
					_, err := cdn.NewDirectoryFrom("")

					So(err, ShouldNotBeNil)
				})

				Convey("When ':8080'", func() {
					_, err := cdn.NewDirectoryFrom(":8080")

					So(err, ShouldNotBeNil)
				})

				Convey("When '@test'", func() {
					_, err := cdn.NewDirectoryFrom("@test")

					So(err, ShouldNotBeNil)
				})
			})
		})
	})
}
