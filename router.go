package water

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
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

// --- route ---

type route struct {
	method, uri string
	handlers    []Handler
}

// routeStore represents a thread-safe store for route pattern.
// 用于检查route pattern是否重复(冲突)及以后打印
type routeStore struct {
	routeMap   map[string]map[string]*route
	routeSlice []*route

	lock sync.Mutex
}

func newRouteStore() *routeStore {
	rs := &routeStore{
		routeMap:   make(map[string]map[string]*route),
		routeSlice: make([]*route, 0),
	}

	for m := range _HTTP_METHODS {
		rs.routeMap[m] = make(map[string]*route)
	}

	return rs
}

func (rs *routeStore) add(r *route) {
	rs.lock.Lock()
	defer rs.lock.Unlock()

	if rs.routeMap[r.method][r.uri] != nil {
		panic(fmt.Sprintf("double uri : %s[%s]", r.method, r.uri))
	}

	rs.routeMap[r.method][r.uri] = r
	rs.routeSlice = append(rs.routeSlice, r)
}

// --- router ---

// 树(单父多子)
type Router struct {
	method  string // 只有终端节点有
	pattern string

	befores  []interface{}
	handlers []interface{} // 只有终端节点有

	parent *Router
	sub    []*Router
}

func NewRouter() *Router {
	return new(Router)
}

func (r *Router) Group(pattern string, fn func(*Router)) {
	rr := &Router{
		pattern: pattern,
		parent:  r,
	}

	r.sub = append(r.sub, rr)

	fn(rr)
}

func (r *Router) Use(handlers ...interface{}) {
	r.befores = append(r.befores, handlers...)
}

func dumpRoute(r *Router, rs *routeStore) {
	if r.sub == nil {
		rs.add(getRoute(r))
		return
	}

	for _, v := range r.sub {
		dumpRoute(v, rs)
	}
}

func getRoute(r *Router) *route {
	ps := []string{}
	hs := []interface{}{}

	tmp := r
	for {
		ps = append(ps, strings.TrimSpace(tmp.pattern))

		if len(tmp.handlers) > 0 {
			hs = append(hs, tmp.handlers...)
		}
		if len(tmp.befores) > 0 {
			hstmp := make([]interface{}, len(tmp.befores))

			copy(hstmp, tmp.befores)
			hstmp = append(hstmp, hs...)
			hs = hstmp
		}

		if tmp.parent == nil {
			break
		}

		tmp = tmp.parent
	}

	re := &route{
		method:   r.method,
		uri:      strings.Join(reverseStrings(ps), ""),
		handlers: newHandlers(hs),
	}

	if len(re.handlers) == 0 {
		panic(fmt.Sprintf("handler err : empty handlers in route(%s,%s)", re.method, re.uri))
	}

	return re
}

// r is root router.
func (r *Router) Handler() *water {
	if r.parent != nil {
		panic("sub router not allowed: Handler")
	}

	rs := newRouteStore()

	dumpRoute(r, rs)

	w := newWater()
	w.routeStore = rs
	w.buildTree()

	return w
}

// 此时还无法获取Router.afters,因为Router.afters还未执行到
// RouteStor放入Route后才可获取afters的信息
func (r *Router) handle(method, pattern string, handlers []interface{}) {
	for _, v := range handlers {
		if v == nil {
			panic(fmt.Sprintf("handler err : find nil in route(%s,%s)", method, pattern))
		}
	}

	rr := &Router{
		method:   method,
		pattern:  pattern,
		parent:   r,
		handlers: handlers,
	}

	r.sub = append(r.sub, rr)
}

func (r *Router) Any(pattern string, handlers ...interface{}) {
	for m := range _HTTP_METHODS {
		r.handle(m, pattern, handlers)
	}
}

func (r *Router) Get(pattern string, handlers ...interface{}) {
	r.handle(http.MethodGet, pattern, handlers)
}

func (r *Router) Post(pattern string, handlers ...interface{}) {
	r.handle(http.MethodPost, pattern, handlers)
}

func (r *Router) Put(pattern string, handlers ...interface{}) {
	r.handle(http.MethodPut, pattern, handlers)
}

func (r *Router) Patch(pattern string, handlers ...interface{}) {
	r.handle(http.MethodPatch, pattern, handlers)
}

func (r *Router) Delete(pattern string, handlers ...interface{}) {
	r.handle(http.MethodDelete, pattern, handlers)
}

func (r *Router) Options(pattern string, handlers ...interface{}) {
	r.handle(http.MethodOptions, pattern, handlers)
}

func (r *Router) Head(pattern string, handlers ...interface{}) {
	r.handle(http.MethodHead, pattern, handlers)
}

func (r *Router) Trace(pattern string, handlers ...interface{}) {
	r.handle(http.MethodTrace, pattern, handlers)
}

func (r *Router) Classic() {
	if r.parent != nil {
		panic("sub router not allowed : Classic()")
	}

	r.Use(Logger())
	r.Use(Recovery())
}
