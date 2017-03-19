# water

water is a micro & pluggable web framework for Go.

> Routing Policy is from [Macaron](github.com/go-macaron/macaron). Thanks [Unknwon](https://github.com/Unknwon).

## Getting Started

Please use latest go.

To install water:

	go get github.com/meilihao/water

The very basic usage of water:

```go
package main

import (
	"fmt"
	"log"

	"github.com/meilihao/water"
)

func main() {
	router := water.NewRouter()

	router.Use(middleware)

	router.Get("/", test)
	router.Get("/help", test)
	router.Any("/about", test)

	router.Group("/a", func(r *water.Router) {
		r.Use(middleware)

		r.Get("/1", test)
		r.Get("/<id:int>", test2)
		r.Group("/b", func(r *water.Router) {
			r.Get("", test, test)
			r.Any("/2", test)
			r.Get("/<id ~ 70|80>", test2)
			r.Get("/*", test)
		})
	})
	router.Get("/d2/<id ~ z(d*)b>", test3)
	router.Get("/d2/<id1+id2 ~ z(d*)h(u)b>", test3)

	w := router.Handler()
	w.PrintRoutes("GET")

	fmt.Println("---###---")

	w.PrintTree("GET")

	if err := water.ListenAndServe(":8080", w); err != nil {
		log.Fatalln(err)
	}
}

func middleware(ctx *water.Context) {
	fmt.Println("1")

	ctx.Next()

	fmt.Println("2")
}

func test(ctx *water.Context) {
	ctx.WriteString(ctx.Request.RequestURI)
}

func test2(ctx *water.Context) {
	ctx.WriteJson(ctx.Params.MustInt("id"))
}

func test3(ctx *water.Context) {
	ctx.WriteJson(ctx.Params)
}
```

## Middlewares

Middlewares allow you easily plugin/unplugin features for your water applications.

There are already some middlewares to simplify your work:

- logger
- recovery
- [cache](https://github.com/meilihao/water-contrib/tree/master/cache) : [cache-memory](https://github.com/meilihao/water-contrib/tree/master/cache),[cache-ssdb](https://github.com/meilihao/water-contrib/tree/master/cache/ssdb)

## Getting Help

- [API Reference](https://gowalker.org/github.com/meilihao/water)

## License

This project is under BSD License.
