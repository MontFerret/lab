package testing_test

import (
	t "testing"

	. "github.com/smartystreets/goconvey/convey"

	testing2 "github.com/MontFerret/lab/pkg/testing"
)

func TestParams(t *t.T) {
	Convey("Params", t, func() {
		Convey(".ToMap", func() {
			Convey("Should respect serialization tags", func() {
				params := testing2.NewParams()
				params.SetSystemValue("data", testing2.DataContext{
					Query: testing2.DataContextValues{
						Result: map[string]interface{}{
							"Foo": "Bar",
						},
						Params: make(map[string]interface{}),
					},
				})

				m := params.ToMap()

				So(m, ShouldResemble, map[string]interface{}{
					"lab": map[string]interface{}{
						"data": map[string]interface{}{
							"query": map[string]interface{}{
								"result": map[string]interface{}{
									"Foo": "Bar",
								},
								"params": make(map[string]interface{}),
							},
						},
					},
				})
			})
		})
	})
}
