package water

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	_HTTP_METHODS = map[string]int{
		"GET":     0,
		"POST":    1,
		"DELETE":  2,
		"PUT":     3,
		"PATCH":   4,
		"HEAD":    5,
		"OPTIONS": 6,
		"TRACE":   7,
	}
)

func methodIndex(method string) int {
	switch method {
	case "GET":
		return 0
	case "POST":
		return 1
	case "DELETE":
		return 2
	case "PUT":
		return 3
	case "PATCH":
		return 4
	case "HEAD":
		return 5
	case "OPTIONS":
		return 6
	case "TRACE":
		return 7
	default:
		return -1
	}
}

type Router struct {
	routers [8]*Tree
	befores []interface{}
	*routeMap
	*groupMap
	logger            *log.Logger
	serial            SerialAdapter
	ctxPool           sync.Pool
	RedirectFixedPath bool
}

func NewRouter() *Router {
	return NewRouterWithLogger(os.Stdout)
}

func NewRouterWithLogger(out io.Writer) *Router {
	r := &Router{
		routers:           [8]*Tree{},
		befores:           make([]interface{}, 0),
		routeMap:          newRouteMap(),
		groupMap:          newGroupMap(),
		logger:            log.New(out, "", 0),
		serial:            nil,
		RedirectFixedPath: true,
	}

	r.ctxPool.New = func() interface{} {
		return newContext(r)
	}

	return r
}

func Classic() *Router {
	r := NewRouter()
	r.Before(Logger())
	r.Before(Recovery())

	return r
}

func (r *Router) SetSerialAdapter(sa SerialAdapter) {
	r.serial = sa
}

func (r *Router) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if !req.ProtoAtLeast(1, 1) || req.RequestURI == "*" || req.Method == "CONNECT" {
		r.log(http.StatusBadRequest, req)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	if r.RedirectFixedPath {
		if p := cleanPath(req.URL.Path); p != req.URL.Path {
			u := *req.URL
			u.Path = p
			r.log(http.StatusMovedPermanently, req)
			http.Redirect(rw, req, u.String(), http.StatusMovedPermanently)
			return
		}
	}

	index := methodIndex(req.Method)
	if index < 0 {
		r.log(http.StatusBadRequest, req)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	handlerChain, params, ok := r.routers[index].Match(req.URL.Path)
	if !ok {
		r.log(http.StatusNotFound, req)
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	ctx := r.ctxPool.Get().(*Context)

	ctx.reset()

	ctx.Environ = make(Environ)
	ctx.Params = params
	ctx.ResponseWriter = rw.(ResponseWriter)
	ctx.Req = req
	ctx.handlers = handlerChain
	ctx.handlersLength = len(handlerChain)

	if r.serial != nil {
		ctx.Id = r.serial.Id()
	}

	ctx.run()

	r.ctxPool.Put(ctx)
}

// handle log before invoke Logger()
// 处理调用Logger()前的日志
func (r *Router) log(status int, req *http.Request) {
	if LogClose {
		return
	}

	start := time.Now()
	r.logger.Printf("%s %v |%s| %13v | %16s | %7s %s\n",
		"[ water ]",
		start.Format(LogTimeFormat),
		logStatus(status),
		time.Now().Sub(start),
		requestRemoteIp(req),
		req.Method,
		req.URL.String(),
	)
}

func (r *Router) Before(handlers ...interface{}) {
	r.befores = append(r.befores, handlers...)
}

func (r *Router) Any(pattern string, handlers ...interface{}) {
	r.handle("Any", pattern, handlers)
}

func (r *Router) Get(pattern string, handlers ...interface{}) {
	r.handle("GET", pattern, handlers)
}

func (r *Router) Post(pattern string, handlers ...interface{}) {
	r.handle("POST", pattern, handlers)
}

func (r *Router) Delete(pattern string, handlers ...interface{}) {
	r.handle("DELETE", pattern, handlers)
}

func (r *Router) Put(pattern string, handlers ...interface{}) {
	r.handle("PUT", pattern, handlers)
}

func (r *Router) Patch(pattern string, handlers ...interface{}) {
	r.handle("PATCH", pattern, handlers)
}

func (r *Router) Options(pattern string, handlers ...interface{}) {
	r.handle("OPTIONS", pattern, handlers)
}

func (r *Router) Head(pattern string, handlers ...interface{}) {
	r.handle("HEAD", pattern, handlers)
}

func (r *Router) Trace(pattern string, handlers ...interface{}) {
	r.handle("TRACE", pattern, handlers)
}

func (r *Router) handle(method, pattern string, handlers []interface{}) {
	if !(pattern == "/" || checkSplitPattern(pattern)) {
		panic(fmt.Sprintf("invalid r.%s pattern : [%s]", method, pattern))
	}

	if _, ok := _HTTP_METHODS[method]; !(ok || method == "Any") {
		panic("unknown HTTP method: " + method)
	}

	methods := make(map[string]bool)
	if method == "Any" {
		for m := range _HTTP_METHODS {
			methods[m] = true
		}
	} else {
		methods[method] = true
	}

	for m := range methods {
		if r.routeMap.isExist(m, pattern) {
			panic(fmt.Sprintf("double pattern : %s[%s]", m, pattern))
		}
	}

	tmpHandlers := make([]interface{}, 0)
	if len(r.befores) > 0 {
		tmpHandlers = append(tmpHandlers, r.befores...)
		tmpHandlers = append(tmpHandlers, handlers...)
	} else {
		tmpHandlers = handlers
	}
	routeHandlers := newHandlers(tmpHandlers)

	for m := range methods {
		if t := r.routers[methodIndex(m)]; t != nil {
			t.Add(pattern, routeHandlers)
		} else {
			t := NewTree()
			t.Add(pattern, routeHandlers)
			r.routers[methodIndex(m)] = t
		}
		r.routeMap.add(m, pattern)
	}
}

func NewHandler(handler interface{}) Handler {
	switch t := handler.(type) {
	case Handler:
		return t
	case func(*Context):
		return HandlerFunc(t)
	/*case http.Handler:
		return HandlerFunc(func(ctx *Context) {
			t.ServeHTTP(ctx.Resp, ctx.Req)
		})
	case func(http.ResponseWriter, *http.Request):
		return HandlerFunc(func(ctx *Context) {
			t(ctx.Resp, ctx.Req)
		})*/
	default:
		panic("invalid handler")
	}
}

func newHandlers(handlers []interface{}) []Handler {
	hs := make([]Handler, len(handlers))
	for k, v := range handlers {
		hs[k] = NewHandler(v)
	}
	return hs
}

// routeMap represents a thread-safe map for route pattern.
// 用于检查route pattern是否重复(冲突)
type routeMap struct {
	routes map[string]map[string]bool
	lock   sync.RWMutex
}

func newRouteMap() *routeMap {
	rm := &routeMap{
		routes: make(map[string]map[string]bool),
	}

	for m := range _HTTP_METHODS {
		rm.routes[m] = make(map[string]bool)
	}

	return rm
}

func (rm *routeMap) isExist(method, pattern string) bool {
	rm.lock.RLock()
	defer rm.lock.RUnlock()

	return rm.routes[method][pattern]
}

func (rm *routeMap) add(method, pattern string) {
	rm.lock.Lock()
	defer rm.lock.Unlock()

	rm.routes[method][pattern] = true
}

func checkMethod(method string) (string, int) {
	method = strings.ToUpper(method)
	idx := methodIndex(method)
	if idx < 0 {
		panic("unknown Support method: " + method)
	}

	return method, idx
}

// print routes by method
// 打印指定方法的路由
func (r *Router) PrintRoutes(method string) {
	method, _ = checkMethod(method)
	routes := r.routeMap.routes[method]

	list := make([]string, 0, len(routes))
	for k := range routes {
		list = append(list, k)
	}

	sort.Strings(list)

	for _, v := range list {
		fmt.Println(v)
	}
}

// print router tree by method
// 打印指定方法的路由树
func (r *Router) PrintTree(method string) {
	_, idx := checkMethod(method)
	tree := r.routers[idx]

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
