package testing_test

import (
	t "testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/MontFerret/lab/testing"
)

func TestHelpers(t *t.T) {
	Convey("ToMap", t, func() {
		Convey("Should copy plain map", func() {
			original := map[string]interface{}{
				"foo": map[string]interface{}{
					"bar": "baz",
				},
			}
			copied := testing.ToMap(original)

			So(&original, ShouldNotPointTo, &copied)
			So(original, ShouldNotEqual, copied)
			So(original, ShouldResemble, copied)
		})

		Convey("Should copy structs", func() {
			type TestStruct struct {
				Foo string
				Bar int
			}

			stct := TestStruct{
				Foo: "bar",
				Bar: 1,
			}

			original := map[string]interface{}{
				"test": stct,
			}
			copied := testing.ToMap(original)

			So(&original, ShouldNotPointTo, &copied)
			So(original, ShouldNotEqual, copied)
			So(original, ShouldNotResemble, copied)
			So(copied["test"].(map[string]interface{})["Foo"], ShouldEqual, stct.Foo)
			So(copied["test"].(map[string]interface{})["Bar"], ShouldEqual, stct.Bar)
		})

		Convey("Should respect struct's json tags", func() {
			type NestedStruct struct {
				Baz []byte `json:"baz"`
			}

			type TestStruct struct {
				Foo    string       `json:"foo"`
				Bar    int          `json:"bar"`
				Nested NestedStruct `json:"nested"`
			}

			stct := TestStruct{
				Foo: "bar",
				Bar: 1,
			}

			original := map[string]interface{}{
				"test": stct,
			}
			copied := testing.ToMap(original)

			So(&original, ShouldNotPointTo, &copied)
			So(original, ShouldNotEqual, copied)
			So(original, ShouldNotResemble, copied)
			So(copied["test"].(map[string]interface{})["foo"], ShouldEqual, stct.Foo)
			So(copied["test"].(map[string]interface{})["bar"], ShouldEqual, stct.Bar)
			So(copied["test"].(map[string]interface{})["nested"].(map[string]interface{})["baz"], ShouldEqual, stct.Nested.Baz)
		})
	})
}
