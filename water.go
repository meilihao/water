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
	rootRouter        *Router
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

	index := MethodIndex(req.Method)
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
		"[ water ]",
		logStatus(status),
		time.Now().Sub(start),
		requestRemoteIp(req),
		req.Method,
		req.URL.String(),
	)
}
