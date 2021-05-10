package water

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestWithNoFoundHandlers(t *testing.T) {
	Convey("WithNoFoundHandlers", t, func() {
		mw := WrapHandler(func(c *Context) {

		})

		r := NewRouter()
		r.Use(mw)
		r.GET("/a", func(c *Context) {

		})
		e := r.Handler(WithNoFoundHandlers(mw))

		{
			resp := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "http://localhost:8080/a", nil)
			So(err, ShouldBeNil)
			e.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, http.StatusOK)
		}

		{
			resp := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "http://localhost:8080/docs/openapi-ui", nil)
			So(err, ShouldBeNil)
			e.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, http.StatusOK)
		}

		{
			resp := httptest.NewRecorder()
			req, err := http.NewRequest("POST", "http://localhost:8080/docs/openapi", nil)
			So(err, ShouldBeNil)
			e.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, http.StatusOK)
		}
	})
}
