package water

import (
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
	noRouteHandler Handler
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

// SetNoFoundHandler the handler for no match route
// for vue spa
func (e *Engine) SetNoFoundHandler(h Handler) {
	e.noRouteHandler = h
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

	ctx := e.ctxPool.Get().(*Context)
	ctx.reset()

	// fast match for static routes
	if ctx.endNode = e.routersStatic[index][req.URL.Path]; ctx.endNode != nil {
	} else {
		// curl http://localhost:8081 or http://localhost:8081/ -> req.URL.Path=="/"
		ctx.endNode, ctx.Params = e.routers[index].Match(req.URL.Path)
	}

	if ctx.endNode == nil {
		if e.noRouteHandler != nil {
			e.noRouteHandler.ServeHTTP(ctx)
		} else {
			ctx.WriteHeader(http.StatusNotFound)
		}

		e.ctxPool.Put(ctx)
		return
	}

	ctx.Environ = make(Environ)

	ctx.ResponseWriter = rw.(ResponseWriter)
	ctx.Request = req

	ctx.handlers = ctx.endNode.handlers
	ctx.handlersLength = len(ctx.handlers)

	ctx.run()

	e.ctxPool.Put(ctx)
}

// Run start web service
// Deprecated: please use Run()
func (e *Engine) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, e)
}

// Run start web service with tls
// Deprecated: please use RunTLS()
func (e *Engine) ListenAndServeTLS(addr, certFile, keyFile string) error {
	return http.ListenAndServeTLS(addr, certFile, keyFile, e)
}

// Run start web service
// defualt port is ":8080"
func (e *Engine) Run(addr ...string) error {
	wantAddr := resolveAddress(addr)

	return http.ListenAndServe(wantAddr, e)
}

// Run start web service with tls
func (e *Engine) RunTLS(addr, certFile, keyFile string) error {
	return http.ListenAndServeTLS(addr, certFile, keyFile, e)
}

func (e *Engine) buildTree() {
	var endNode *node

	for _, v := range e.routeStore.routeSlice {
		if t := e.routers[MethodIndex(v.method)]; t != nil {
			endNode = t.add(v.variantUri, v.handlers)
		} else {
			t := newTree()
			endNode = t.add(v.variantUri, v.handlers)
			e.routers[MethodIndex(v.method)] = t
		}

		if isStaticRoute(endNode) {
			if e.routersStatic[MethodIndex(v.method)] == nil {
				e.routersStatic[MethodIndex(v.method)] = map[string]*node{}
			}
			e.routersStatic[MethodIndex(v.method)][v.variantUri] = endNode
		}

		endNode.matchNode = v
	}
}

// 向上递归检查是否为static route
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
