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
	router := water.Classic()

	router.Get("/", test)
	router.Get("/help", test)
	router.Any("/about", test)

	router.Group("/a", func(g *water.Group) {
		g.Before(middleware)

		g.Get("/1", test)
		g.Get("/<id:int>", test2)
		g.Group("/b", func(g *water.Group) {
			g.Get("", test)
			g.Any("/2", test)
			g.Get("/<id ~ 70|80>", test2)
			g.Get("/*", test)
		})
	})

	router.PrintRoutes("GET")

	fmt.Println("---###---")

	router.PrintTree("GET")

	if err := router.ListenAndServe(":8080"); err != nil {
		log.Fatalln(err)
	}
}

func middleware(ctx *water.Context) {
	fmt.Println("1")

	ctx.Next()

	fmt.Println("2")
}

func test(ctx *water.Context) {
	ctx.WriteString(ctx.Req.RequestURI)
}

func test2(ctx *water.Context) {
	ctx.WriteJson(ctx.Params.MustInt("id"))
}
```

## Middlewares

Middlewares allow you easily plugin/unplugin features for your water applications.

There are already some middlewares to simplify your work:

- logger
- recovery

## Getting Help

- [API Reference](https://gowalker.org/github.com/meilihao/water)

## License

This project is under BSD License.