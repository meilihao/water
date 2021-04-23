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
		tmp := _VariantUri(raw)

		So(tmp, ShouldEqual, tmp)
	})
}

func TestMatchHolder(t *testing.T) {
	Convey("with 0:", t, func() {
		t := newTree()
		t.add("/a/<id>", nil)
		t.add("/b/<_>", nil)

		end, params := t.Match("/a/1")
		So(end, ShouldNotBeNil)
		So(len(params), ShouldEqual, 1)
		So(params["id"], ShouldEqual, "1")

		end, params = t.Match("/b/2")
		So(end, ShouldNotBeNil)
		So(len(params), ShouldEqual, 0)
	})
}

func TestMatchAll(t *testing.T) {
	Convey("with 0:", t, func() {
		t := newTree()
		t.add("/*", nil)

		end, params := t.Match("/a.png")
		So(end, ShouldNotBeNil)
		So(len(params), ShouldEqual, 1)
		So(params["*0"], ShouldEqual, "a.png")

		end, params = t.Match("/f/a.png")
		So(end, ShouldNotBeNil)
		So(len(params), ShouldEqual, 1)
		So(params["*0"], ShouldEqual, "f/a.png")
	})

	Convey("with 1 no holder:", t, func() {
		t := newTree()
		t.add("/file/*", nil)

		end, params := t.Match("/file/a.png")
		So(end, ShouldNotBeNil)
		So(len(params), ShouldEqual, 1)
		So(params["*0"], ShouldEqual, "a.png")
	})

	Convey("with 1 with holder:", t, func() {
		t := newTree()
		t.add("/file/*f", nil)

		end, params := t.Match("/file/a.png")
		So(end, ShouldNotBeNil)
		So(len(params), ShouldEqual, 1)
		So(params["f"], ShouldEqual, "a.png")
	})

	Convey("with 1 ignore holder:", t, func() {
		t := newTree()
		t.add("/file/*_", nil)

		end, params := t.Match("/file/a.png")
		So(end, ShouldNotBeNil)
		So(len(params), ShouldEqual, 0)
	})

	Convey("with 2 no holder:", t, func() {
		t := newTree()
		t.add("/file/*/*", nil)

		end, params := t.Match("/file/f/a.png")
		So(end, ShouldNotBeNil)
		So(len(params), ShouldEqual, 2)
		So(params["*0"], ShouldEqual, "f")
		So(params["*1"], ShouldEqual, "a.png")
	})

	Convey("with 2 with holder:", t, func() {
		t := newTree()
		t.add("/file/*/*f", nil)

		end, params := t.Match("/file/f/a.png")
		So(end, ShouldNotBeNil)
		So(len(params), ShouldEqual, 2)
		So(params["f"], ShouldEqual, "a.png")
	})

	Convey("with 1 ignore holder:", t, func() {
		t := newTree()
		t.add("/file/*/*_", nil)

		end, params := t.Match("/file/f/a.png")
		So(end, ShouldNotBeNil)
		So(len(params), ShouldEqual, 1)
	})
}
