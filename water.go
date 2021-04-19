package water

import (
	"fmt"
	"net/http"
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
	default:
		panic("unsupported handler")
	}
}

func newHandlers(handlers []interface{}) (a []Handler) {
	n := len(handlers)

	a = make([]Handler, n)
	for i, h := range handlers {
		a[i] = newHandler(h)
	}

	return a
}

// BeforeHandler represents a handler executes at beginning of every request(before HandlerFuncs).
// Water stops future process when it returns true.
type BeforeHandler func(http.ResponseWriter, *http.Request) bool

// --- water ---
type Engine struct {
	rootRouter    *Router
	routers       [8]*node
	routersStatic [8]map[string]*node
	routeStore    *routeStore
	ctxPool       sync.Pool

	// BeforeHandlers []BeforeHandler
}

func newWater() *Engine {
	e := &Engine{
		routers:       [8]*node{},
		routersStatic: [8]map[string]*node{},
	}

	e.ctxPool.New = func() interface{} {
		return newContext()
	}

	return e
}

func (e *Engine) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if !req.ProtoAtLeast(1, 1) || req.RequestURI == "*" || req.Method == "CONNECT" {
		rw.WriteHeader(http.StatusNotAcceptable)
		return
	}

	index := MethodIndex(req.Method)
	if index < 0 {
		rw.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// if len(e.BeforeHandlers) > 0 {
	// 	for _, h := range e.BeforeHandlers {
	// 		if h(rw, req) {
	// 			return
	// 		}
	// 	}
	// }

	var handlerChain []Handler
	var params Params
	var found bool

	// fast match for static routes
	if node := e.routersStatic[index][req.URL.Path]; node != nil {
		handlerChain = node.handlers
		found = true
	} else {
		// curl http://localhost:8081 or http://localhost:8081/ -> req.URL.Path=="/"
		handlerChain, params, found = e.routers[index].Match(req.URL.Path)
	}

	if !found {
		e.log(http.StatusNotFound, req)
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	ctx := e.ctxPool.Get().(*Context)

	ctx.reset()

	ctx.Environ = make(Environ)
	ctx.Params = params
	ctx.ResponseWriter = rw.(ResponseWriter)
	ctx.Request = req
	ctx.handlers = handlerChain
	ctx.handlersLength = len(handlerChain)

	ctx.run()

	e.ctxPool.Put(ctx)
}

func (e *Engine) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, e)
}

func (e *Engine) ListenAndServeTLS(addr, certFile, keyFile string) error {
	return http.ListenAndServeTLS(addr, certFile, keyFile, e)
}

func (e *Engine) buildTree() {
	var endNode *node

	for _, v := range e.routeStore.routeSlice {
		if !(v.uri == "/" || checkSplitPattern(v.uri)) {
			panic(fmt.Sprintf("invalid route : [%s : %s]", v.method, v.uri))
		}

		if t := e.routers[MethodIndex(v.method)]; t != nil {
			endNode = t.add(v.uri, v.handlers)
		} else {
			t := newTree()
			endNode = t.add(v.uri, v.handlers)
			e.routers[MethodIndex(v.method)] = t
		}

		if isStaticRoute(endNode) {
			if e.routersStatic[MethodIndex(v.method)] == nil {
				e.routersStatic[MethodIndex(v.method)] = map[string]*node{}
			}
			e.routersStatic[MethodIndex(v.method)][v.uri] = endNode
		}
	}
}

func isStaticRoute(node *node) bool {
	if node == nil {
		return true
	}

	if node.typ != _PATTERN_STATIC {
		return false
	}

	return isStaticRoute(node.parent)
}

// handle log before invoke Logger()
// 处理调用Logger()前的日志
func (e *Engine) log(status int, req *http.Request) {
	if LogClose {
		return
	}

	start := time.Now()
	logx.Infof("%s |%s| %13v | %16s | %7s %s",
		logPrefix(req),
		logStatus(status),
		time.Now().Sub(start),
		requestRealIp(req),
		req.Method,
		req.URL.String(),
	)
}
