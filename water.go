package water

import (
	"fmt"
	"net/http"
	"sort"
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

// support http.Handler
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
	routers           [8]*Tree
	routeStore        *routeStore
	serial            SerialAdapter
	ctxPool           sync.Pool
	RedirectFixedPath bool
}

func newWater() *water {
	w := &water{
		routers:           [8]*Tree{},
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
	ctx.Req = req
	ctx.handlers = handlerChain
	ctx.handlersLength = len(handlerChain)

	if w.serial != nil {
		ctx.Id = w.serial.Id()
	}

	ctx.run()

	w.ctxPool.Put(ctx)
}

func (w *water) BuildTree() {
	for _, v := range w.routeStore.routeSlice {
		if !(v.uri == "/" || checkSplitPattern(v.uri)) {
			panic(fmt.Sprintf("invalid r.%s pattern : [%s]", v.method, v.uri))
		}

		if t := w.routers[methodIndex(v.method)]; t != nil {
			t.Add(v.uri, v.handlers)
		} else {
			t := NewTree()
			t.Add(v.uri, v.handlers)
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
	logx.Infof("%s |%s| %13v | %16s | %7s %s\n",
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

// print router tree by method
// 打印指定方法的路由树
// TODO 打印[]handler的名称
func (w *water) PrintTree(method string) {
	_, idx := checkMethod(method)
	tree := w.routers[idx]

	fmt.Println("/")
	printTreeNode(0, tree)
}

func printTreeNode(depth int, tree *Tree) {
	space := "│"
	currentTree := tree
	for {
		n := len(currentTree.pattern)
		if currentTree.parent != nil {
			for i := 0; i < n; i++ {
				space += " "
			}
			currentTree = currentTree.parent
		} else {
			break
		}
	}
	for i := 0; i < depth*3; i++ { //每层"── "的宽度
		space += " "
	}

	// the same order with tree.matchSubtree
	if len(tree.subtrees) > 0 {
		for _, v := range tree.subtrees {
			fmt.Println(fmt.Sprintf("%s── %s", space, v.pattern))

			printTreeNode(depth+1, v)
		}
	}

	if len(tree.leaves) > 0 {
		for _, v := range tree.leaves {
			fmt.Println(fmt.Sprintf("%s── %s", space, v.pattern))
		}
	}
}
