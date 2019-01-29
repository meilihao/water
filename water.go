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
type water struct {
	rootRouter *Router
	routers    [8]*node
	routeStore *routeStore
	ctxPool    sync.Pool

	BeforeHandlers []BeforeHandler
}

func newWater() *water {
	w := &water{
		routers: [8]*node{},
	}

	w.ctxPool.New = func() interface{} {
		return newContext()
	}

	return w
}

func (w *water) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if !req.ProtoAtLeast(1, 1) || req.RequestURI == "*" || req.Method == "CONNECT" {
		rw.WriteHeader(http.StatusNotAcceptable)
		return
	}

	if len(w.BeforeHandlers)>0 {
		for _, h := range w.BeforeHandlers {
			if h(rw, req) {
				return
			}
		}
	}

	index := MethodIndex(req.Method)
	if index < 0 {
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

	ctx.run()

	w.ctxPool.Put(ctx)
}

func (w *water) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, w)
}

func (w *water) ListenAndServeTLS(addr, certFile, keyFile string) error {
	return http.ListenAndServeTLS(addr, certFile, keyFile, w)
}

func (w *water) buildTree() {
	for _, v := range w.routeStore.routeSlice {
		if !(v.uri == "/" || checkSplitPattern(v.uri)) {
			panic(fmt.Sprintf("invalid route : [%s : %s]", v.method, v.uri))
		}

		if t := w.routers[MethodIndex(v.method)]; t != nil {
			t.add(v.uri, v.handlers)
		} else {
			t := newTree()
			t.add(v.uri, v.handlers)
			w.routers[MethodIndex(v.method)] = t
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
		logPrefix(req),
		logStatus(status),
		time.Now().Sub(start),
		requestRealIp(req),
		req.Method,
		req.URL.String(),
	)
}
