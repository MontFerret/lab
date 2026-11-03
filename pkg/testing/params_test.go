package testing_test

import (
	t "testing"

	. "github.com/smartystreets/goconvey/convey"

	labtesting "github.com/MontFerret/lab/v2/pkg/testing"
)

func TestParams(t *t.T) {
	Convey("Params", t, func() {
		Convey(".ToMap", func() {
			Convey("Should respect serialization tags", func() {
				params := labtesting.NewParams()
				params.SetSystemValue("data", labtesting.DataContext{
					Query: labtesting.DataContextValues{
						Result: map[string]any{
							"Foo": "Bar",
						},
						Params: make(map[string]any),
					},
				})

				m := params.ToMap()

				So(m, ShouldResemble, map[string]any{
					"lab": map[string]any{
						"data": map[string]any{
							"query": map[string]any{
								"result": map[string]any{
									"Foo": "Bar",
								},
								"params": make(map[string]any),
							},
						},
					},
				})
			})
		})
	})
}
