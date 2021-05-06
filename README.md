# water

water is a micro & pluggable web framework for Go.

**water is Compatible with gin route style.**

> Routing Policy is from [Macaron](github.com/go-macaron/macaron). Thanks [Unknwon](https://github.com/Unknwon).

## Getting Started

Please use latest go.

To install water:

	go get github.com/meilihao/water

The very basic usage of water:

example see [engine_test.go](/engine_test.go)

output(router tree):
```sh
 Raw Router Tree:
├── / [GET     : 2]
├── /help [GET     : 2]
├── /about [GET     : 2]
├── /about [POST    : 2]
├── /about [DELETE  : 2]
├── /about [PUT     : 2]
├── /about [PATCH   : 2]
├── /about [HEAD    : 2]
├── /* [OPTIONS : 1]
├── /a
│   ├── /1 [GET     : 3]
│   ├── /<id:int> [GET     : 3]
│   └── /b
│       ├──  [GET     : 4]
│       ├── /2 [GET     : 3]
│       ├── /2 [POST    : 3]
│       ├── /2 [DELETE  : 3]
│       ├── /2 [PUT     : 3]
│       ├── /2 [PATCH   : 3]
│       ├── /<id ~ 70|80> [PUT     : 3]
│       └── /* [GET     : 3]
├── /d2/<id ~ z(d*)b> [GET     : 2]
├── /d2/<id1,id2 ~ z(d*)h(u)b> [GET     : 2]
├── /c
│   ├── /<_ ~ 70|80> [PUT     : 2]
│   ├── /<_> [GET     : 2]
│   └── /*file [GET     : 2]
└── /d
    └── /*_ [GET     : 2]


 GET's Routes:
[  2] /
[  3] /a/1
[  3] /a/<id:int>
[  4] /a/b
[  3] /a/b/*
[  3] /a/b/2
[  2] /about
[  2] /c/*file
[  2] /c/<_>
[  2] /d/*_
[  2] /d2/<id ~ z(d*)b>
[  2] /d2/<id1,id2 ~ z(d*)h(u)b>
[  2] /help


 All Routes:
[GET     : 2] /
[GET     : 2] /help
[GET     : 2] /about
[POST    : 2] /about
[DELETE  : 2] /about
[PUT     : 2] /about
[PATCH   : 2] /about
[HEAD    : 2] /about
[OPTIONS : 1] /*
[GET     : 3] /a/1
[GET     : 3] /a/<id:int>
[GET     : 4] /a/b
[GET     : 3] /a/b/2
[POST    : 3] /a/b/2
[DELETE  : 3] /a/b/2
[PUT     : 3] /a/b/2
[PATCH   : 3] /a/b/2
[PUT     : 3] /a/b/<id ~ 70|80>
[GET     : 3] /a/b/*
[GET     : 2] /d2/<id ~ z(d*)b>
[GET     : 2] /d2/<id1,id2 ~ z(d*)h(u)b>
[PUT     : 2] /c/<_ ~ 70|80>
[GET     : 2] /c/<_>
[GET     : 2] /c/*file
[GET     : 2] /d/*_


 GET's Release Router Tree:
/ [  2]
├── a
│   ├── b
│   │   ├── 2 [  3]
│   │   └── * [  3]
│   ├── 1 [  3]
│   ├── b [  4]
│   └── <id:int> [  3]
├── d2
│   ├── <id ~ z(d*)b> [  2]
│   └── <id1,id2 ~ z(d*)h(u)b> [  2]
├── c
│   ├── <_> [  2]
│   └── *file [  2]
├── d
│   └── *_ [  2]
├── help [  2]
└── about [  2]
```

## Middlewares

Middlewares allow you easily plugin/unplugin features for your water applications.

There are already some middlewares to simplify your work:

- logger
- recovery
- cors
- static
- [cache](https://github.com/meilihao/water-contrib/tree/master/cache) : [cache-memory](https://github.com/meilihao/water-contrib/tree/master/cache),[cache-ssdb](https://github.com/meilihao/water-contrib/tree/master/cache/ssdb)
- [reqestDump](https://github.com/meilihao/water-contrib/tree/master/debug)

## Router
- default Any() exclude ["Head","Options"], but can reset by `water.MethodAnyExclude`

## binding
support gin style binding.

## others
- support http.Handler, but not recommended
- `go build --tags extended`, enable advanced response

## Getting Help

- [API Reference](https://gowalker.org/github.com/meilihao/water)

## License

This project is under BSD License.