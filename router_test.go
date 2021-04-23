package water

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_VariantUri(t *testing.T) {
	Convey("_VariantUri :", t, func() {
		raw := "/hello/:name"
		vuri := "/hello/<name>"

		tmp := _VariantUri(raw)

		So(tmp, ShouldEqual, vuri)
	})

	Convey("_VariantUri *", t, func() {
		raw := "/hello/*filepath"
		vuri := "/hello/*"

		tmp := _VariantUri(raw)

		So(tmp, ShouldEqual, vuri)
	})
}
