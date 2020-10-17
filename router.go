package water

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
)

var (
	_HTTP_METHODS = map[string]int{
		http.MethodGet:     0,
		http.MethodPost:    1,
		http.MethodDelete:  2,
		http.MethodPut:     3,
		http.MethodPatch:   4,
		http.MethodHead:    5,
		http.MethodOptions: 6,
	}
	_HTTP_METHODS_NAMES = []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodDelete,
		http.MethodPut,
		http.MethodPatch,
		http.MethodHead,
		http.MethodOptions,
	}
)

func MethodIndex(method string) int {
	switch method {
	case http.MethodGet:
		return 0
	case http.MethodPost:
		return 1
	case http.MethodDelete:
		return 2
	case http.MethodPut:
		return 3
	case http.MethodPatch:
		return 4
	case http.MethodHead:
		return 5
	case http.MethodOptions:
		return 6
	default:
		return -1
	}
}

// --- route ---

type route struct {
	method, uri string
	handlers    []Handler
}

// routeStore represents a thread-safe store for route uri.
// to check double route uri and to print route uri
type routeStore struct {
	routeMap   map[string]map[string]*route // [http_method][uri]route
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

// multiway tree
type Router struct {
	method  string // only in router leaf
	pattern string

	befores  []interface{}
	handlers []interface{} // only in router leaf

	parent *Router
	sub    []*Router
}

func NewRouter() *Router {
	return &Router{}
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

func (r *Router) handle(method, pattern string, handlers []interface{}) {
	for _, v := range handlers {
		if v == nil {
			panic(fmt.Sprintf("handler err : find nil in handlers(%s,%s)", method, pattern))
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

var (
	MethodAnyExclude = []string{http.MethodHead, http.MethodOptions}
)

func (r *Router) Any(pattern string, handlers ...interface{}) {
Skip:
	for _, method := range _HTTP_METHODS_NAMES {
		for _, v := range MethodAnyExclude {
			if method == v {
				continue Skip
			}
		}
		r.handle(method, pattern, handlers)
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

// to generate router tree.
// r is root router.
func (r *Router) Handler() *water {
	if r.parent != nil {
		panic("sub router not allowed: Handler()")
	}

	rs := newRouteStore()

	dumpRoute(r, rs)

	w := newWater()
	w.rootRouter = r
	w.routeStore = rs
	w.buildTree()

	return w
}

func (r *Router) Classic() {
	if r.parent != nil {
		panic("sub router not allowed : Classic()")
	}

	r.Use(Logger())
	r.Use(Recovery())
}
