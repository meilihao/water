package water

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

var currentRoot, _ = os.Getwd()

func TestStatic(t *testing.T) {
	Convey("Serve static files deepth=1", t, func() {
		r := NewRouter()
		r.Static(&StaticOptions{
			Prefix:     "/f",
			FileSystem: http.Dir(currentRoot),
			Expires:    func() string { return "46" },
		})
		e := r.Handler()

		resp := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "http://localhost:8080/f/version.go", nil)
		So(err, ShouldBeNil)
		e.ServeHTTP(resp, req)
		So(resp.Code, ShouldNotEqual, http.StatusNotFound)
	})

	Convey("Serve static files deepth=2", t, func() {
		r := NewRouter()
		r.Static(&StaticOptions{
			Prefix:     "/f",
			FileSystem: http.Dir(currentRoot),
			Expires:    func() string { return "46" },
		})
		e := r.Handler()

		resp := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "http://localhost:8080/f/benchmark/direct_or_embed_test.go", nil)
		So(err, ShouldBeNil)
		e.ServeHTTP(resp, req)
		So(resp.Code, ShouldNotEqual, http.StatusNotFound)
	})
}
