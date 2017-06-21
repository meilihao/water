package water

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/meilihao/logx"
)

type Handler interface {
	ServeHTTP(*Context)
}

type HandlerFunc func(*Context)

func (f HandlerFunc) ServeHTTP(ctx *Context) {
	f(ctx)
}

// unsupport http.Handler for default
func newHandler(handler interface{}) Handler {
	switch h := handler.(type) {
	case Handler:
		return h
	case func(*Context):
		return HandlerFunc(h)
		/*	case http.Handler:
				return HandlerFunc(func(ctx *Context) {
					h.ServeHTTP(ctx, ctx.Req)
				})
			case func(http.ResponseWriter, *http.Request):
				return HandlerFunc(func(ctx *Context) {
					h(ctx, ctx.Req)
				})
		*/
	default:
		panic("unsupported handler")
	}
}

func newHandlers(handlers []interface{}) (a []Handler) {
	n := len(handlers)
	if n == 0 {
		panic("empty handlers")
	}

	a = make([]Handler, len(handlers))
	for i, h := range handlers {
		a[i] = newHandler(h)
	}

	return a
}

func ListenAndServe(addr string, handler http.Handler) error {
	return http.ListenAndServe(addr, handler)
}

func ListenAndServeTLS(addr, certFile, keyFile string, handler http.Handler) error {
	return http.ListenAndServeTLS(addr, certFile, keyFile, handler)
}

// --- water ---

type water struct {
	routers           [8]*node
	routeStore        *routeStore
	serial            SerialAdapter
	ctxPool           sync.Pool
	RedirectFixedPath bool
}

func newWater() *water {
	w := &water{
		routers:           [8]*node{},
		serial:            nil,
		RedirectFixedPath: true,
	}

	w.ctxPool.New = func() interface{} {
		return newContext()
	}

	return w
}

func (w *water) SetSerialAdapter(sa SerialAdapter) {
	w.serial = sa
}

func (w *water) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if !req.ProtoAtLeast(1, 1) || req.RequestURI == "*" || req.Method == "CONNECT" {
		w.log(http.StatusBadRequest, req)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	if w.RedirectFixedPath {
		if p := cleanPath(req.URL.Path); p != req.URL.Path {
			u := *req.URL
			u.Path = p
			w.log(http.StatusMovedPermanently, req)
			http.Redirect(rw, req, u.String(), http.StatusMovedPermanently)
			return
		}
	}

	index := methodIndex(req.Method)
	if index < 0 {
		w.log(http.StatusMethodNotAllowed, req)
		rw.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	handlerChain, params, ok := w.routers[index].Match(req.URL.Path)
	if !ok {
		w.log(http.StatusNotFound, req)
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	ctx := w.ctxPool.Get().(*Context)

	ctx.reset()

	ctx.Environ = make(Environ)
	ctx.Params = params
	ctx.ResponseWriter = rw.(ResponseWriter)
	ctx.Request = req
	ctx.handlers = handlerChain
	ctx.handlersLength = len(handlerChain)

	if w.serial != nil {
		ctx.Id = w.serial.Id()
	}

	ctx.run()

	w.ctxPool.Put(ctx)
}

func (w *water) buildTree() {
	for _, v := range w.routeStore.routeSlice {
		if !(v.uri == "/" || checkSplitPattern(v.uri)) {
			panic(fmt.Sprintf("invalid route : [%s : %s]", v.method, v.uri))
		}

		if t := w.routers[methodIndex(v.method)]; t != nil {
			t.add(v.uri, v.handlers)
		} else {
			t := newTree()
			t.add(v.uri, v.handlers)
			w.routers[methodIndex(v.method)] = t
		}
	}
}

// handle log before invoke Logger()
// 处理调用Logger()前的日志
func (w *water) log(status int, req *http.Request) {
	if LogClose {
		return
	}

	start := time.Now()
	logx.Infof("%s |%s| %13v | %16s | %7s %s",
		"[ water ]",
		logStatus(status),
		time.Now().Sub(start),
		requestRemoteIp(req),
		req.Method,
		req.URL.String(),
	)
}

// print routes by method
// 打印指定方法的路由
func (w *water) PrintRoutes(method string) {
	method, _ = checkMethod(method)
	routes := w.routeStore.routeMap[method]

	list := make([]string, 0, len(routes))
	for k := range routes {
		list = append(list, k)
	}

	sort.Strings(list)

	for _, v := range list {
		route := routes[v]

		// count(router.handlers) + uri
		fmt.Printf("(%2d) %s\n", len(route.handlers), v)
	}
}

func (w *water) PrintAllRoutes() {
	for _, v := range w.routeStore.routeSlice {
		// count(router.handlers) + uri
		fmt.Printf("(%7s) %s\n", v.method, v.uri)
	}
}

// print router tree by method
// 打印指定方法的路由树
// TODO 打印[]handler的名称
func (w *water) PrintTree(method string) {
	_, idx := checkMethod(method)
	tree := w.routers[idx]

	printName(tree.pattern, 0)
	dumpTree(tree, 0)
}

func dumpTree(n *node, depth int) {
	if len(n.subNodes) > 0 {
		for _, sub := range n.subNodes {
			printName(sub.pattern, depth)
			dumpTree(sub, depth+1)
		}
	}
	if len(n.endNodes) > 0 {
		for _, end := range n.endNodes {
			printName(end.pattern, depth)
		}
	}
}

func printName(name string, depth int) {
	fmt.Printf("%s+---%s\n", strings.Repeat(" ", depth*4), name)
}
