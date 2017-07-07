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
	router.Any("/about", test) // default Any() exclude ["Head","Options","Trace"]
	router.Head("/about", test)
	router.Options("/*")

	router.Group("/a", func(r *water.Router) {
		r.Use(middleware)

		r.Get("/1", test)
		r.Get("/<id:int>", test2)
		r.Group("/b", func(r *water.Router) {
			r.Get("", test, test)
			r.Any("/2", test)
			r.Put("/<id ~ 70|80>", test2)
			r.Get("/*", test)
		})
	})
	router.Get("/d2/<id ~ z(d*)b>", test3)
	router.Get("/d2/<id1+id2 ~ z(d*)h(u)b>", test3)

	w := router.Handler()

	fmt.Println("\n\n", "Raw Router Tree:")
	w.PrintRawRouter()
	fmt.Println("\n\n", "GET's Routes:")
	w.PrintRawRoutes("GET")
	fmt.Println("\n\n", "All Routes:")
	w.PrintRawAllRoutes()
	fmt.Println("\n\n", "GET's Release Router Tree:")
	w.PrintRouterTree("GET")

	if err := water.ListenAndServe(":8081", w); err != nil {
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

output(router tree):
```sh
 Raw Router Tree:
Routers:
├── / [GET     : 1]
├── /help [GET     : 1]
├── /about [GET     : 1]
├── /about [POST    : 1]
├── /about [DELETE  : 1]
├── /about [PUT     : 1]
├── /about [PATCH   : 1]
├── /about [HEAD    : 1]
├── /* [OPTIONS : 0]
├── /a
│   ├── /1 [GET     : 1]
│   ├── /<id:int> [GET     : 1]
│   └── /b
│       ├──  [GET     : 2]
│       ├── /2 [GET     : 1]
│       ├── /2 [POST    : 1]
│       ├── /2 [DELETE  : 1]
│       ├── /2 [PUT     : 1]
│       ├── /2 [PATCH   : 1]
│       ├── /<id ~ 70|80> [PUT     : 1]
│       └── /* [GET     : 1]
├── /d2/<id ~ z(d*)b> [GET     : 1]
└── /d2/<id1+id2 ~ z(d*)h(u)b> [GET     : 1]


 GET's Routes:
( 2) /
( 3) /a/1
( 3) /a/<id:int>
( 4) /a/b
( 3) /a/b/*
( 3) /a/b/2
( 2) /about
( 2) /d2/<id ~ z(d*)b>
( 2) /d2/<id1+id2 ~ z(d*)h(u)b>
( 2) /help


 All Routes:
(    GET) /
(    GET) /help
(    GET) /about
(   POST) /about
( DELETE) /about
(    PUT) /about
(  PATCH) /about
(   HEAD) /about
(OPTIONS) /*
(    GET) /a/1
(    GET) /a/<id:int>
(    GET) /a/b
(    GET) /a/b/2
(   POST) /a/b/2
( DELETE) /a/b/2
(    PUT) /a/b/2
(  PATCH) /a/b/2
(    PUT) /a/b/<id ~ 70|80>
(    GET) /a/b/*
(    GET) /d2/<id ~ z(d*)b>
(    GET) /d2/<id1+id2 ~ z(d*)h(u)b>


 GET's Release Router Tree:
/ [2]
├── a
│   ├── b
│   │   ├── 2 [3]
│   │   └── * [3]
│   ├── 1 [3]
│   ├── b [4]
│   └── <id:int> [3]
├── d2
│   ├── <id ~ z(d*)b> [2]
│   └── <id1+id2 ~ z(d*)h(u)b> [2]
├── help [2]
└── about [2]
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
