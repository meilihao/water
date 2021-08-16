package water

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEngine(t *testing.T) {
	router := NewRouter()

	router.Use(middleware)

	router.GET("/", test)
	router.GET("/help", test)
	router.GET("/help2", testRaw)
	router.ANY("/about", test)
	router.HEAD("/about", test)
	router.OPTIONS("/*")

	router.Group("/a", func(r *Router) {
		r.Use(middleware)

		r.GET("/1", test)
		r.GET("/<id:int>", test2)
		r.Group("/b", func(r *Router) {
			r.GET("", test, test)
			r.ANY("/2", test)
			r.PUT("/<id ~ 70|80>", test2)
			r.GET("/*", test)
		})
	})
	router.GET("/d2/<id ~ z(d*)b>", test3)
	router.GET("/d2/<id1,id2 ~ z(d*)h(u)b>", test3)
	router.Group("/c", func(r *Router) {
		r.PUT("/<_ ~ 70|80>", test2) // ignore holder
		r.GET("/<_>", test2)
		r.GET("/*file", test) // named match all
	})
	router.Group("/d", func(r *Router) {
		r.GET("/*_", test) // ignore match all
	})

	opts := []Option{
		WithStaticRouter(true),
	}

	w := router.Handler(opts...)

	fmt.Println("\n\n", "Raw Router Tree:")
	w.PrintRawRouter()

	fmt.Println("\n\n", "GET's Routes:")
	w.PrintRawRoutes(http.MethodGet)

	fmt.Println("\n\n", "All Routes:")
	w.PrintRawAllRoutes()

	fmt.Println("\n\n", "GET's Release Router Tree:")
	w.PrintRouterTree(http.MethodGet)

	// if err := w.ListenAndServe(":8081"); err != nil {
	// 	log.Fatalln(err)
	// }
}

func middleware(ctx *Context) {
	fmt.Println("1")

	ctx.Next()

	fmt.Println("2")
}

func test(ctx *Context) {
	ctx.String(200, ctx.Request.RequestURI)
}

func test2(ctx *Context) {
	ctx.JSON(200, ctx.Params.MustInt("id"))
}

func test3(ctx *Context) {
	ctx.JSON(200, ctx.Params)
}

// support http.Handler, but not recommended
func testRaw(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte(req.URL.String()))
}

func TestEngineNoRoute(t *testing.T) {
	router := NewRouter()

	router.Use(middleware)

	e := router.Handler()

	{
		resp := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "http://localhost:8080/a", nil)
		e.ServeHTTP(resp, req)
		if resp.Code != http.StatusNotFound {
			t.Fatalf("want %d, get %d", http.StatusNotFound, resp.Code)
		}
	}

}

func TestNoMethodRoute(t *testing.T) {
	r := NewRouter()
	r.GET("/a", func(c *Context) {

	})
	e := r.Handler()

	{
		resp := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "http://localhost:8080/docs/openapi-ui", nil)
		e.ServeHTTP(resp, req)
		if resp.Code != http.StatusNotFound {
			t.Fatalf("want %d, get %d", http.StatusNotFound, resp.Code)
		}
	}

	// no POST router tree
	{
		resp := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "http://localhost:8080/docs/openapi", nil)
		e.ServeHTTP(resp, req)
		if resp.Code != http.StatusNotFound {
			t.Fatalf("want %d, get %d", http.StatusNotFound, resp.Code)
		}
	}
}
