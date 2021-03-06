package water

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/meilihao/water/binding"
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
	method     string
	uri        string // raw uri
	variantUri string // variant uri, httprouter route compatible
	handlers   []Handler
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

	if r.method == "" { // end route is middleware
		return
	}

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

	gbefores []interface{} // for global middleware, include handle middleware before match routes
	befores  []interface{}
	handlers []interface{} // only in router leaf

	parent *Router
	sub    []*Router
}

func NewRouter() *Router {
	return &Router{}
}

// Before for global middleware, include handle middleware before match routes and 404
func (r *Router) Before(handlers ...interface{}) {
	if !r.IsParent() {
		panic("sub router not allowed: Before()")
	}

	r.gbefores = append(r.gbefores, handlers...)
}

func (r *Router) Group(pattern string, is ...interface{}) *Router {
	rr := &Router{
		pattern: pattern,
		parent:  r,
	}

	r.sub = append(r.sub, rr)

	for i := range is {
		switch v := is[i].(type) {
		case func(*Context):
			r.befores = append(r.befores, v)
		case func(*Router):
			v(rr)
		default:
			panic("unsupported type")
		}
	}

	return rr
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

func (r *Router) ANY(pattern string, handlers ...interface{}) {
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

func (r *Router) GET(pattern string, handlers ...interface{}) {
	r.handle(http.MethodGet, pattern, handlers)
}

func (r *Router) POST(pattern string, handlers ...interface{}) {
	r.handle(http.MethodPost, pattern, handlers)
}

func (r *Router) PUT(pattern string, handlers ...interface{}) {
	r.handle(http.MethodPut, pattern, handlers)
}

func (r *Router) PATCH(pattern string, handlers ...interface{}) {
	r.handle(http.MethodPatch, pattern, handlers)
}

func (r *Router) DELETE(pattern string, handlers ...interface{}) {
	r.handle(http.MethodDelete, pattern, handlers)
}

func (r *Router) OPTIONS(pattern string, handlers ...interface{}) {
	r.handle(http.MethodOptions, pattern, handlers)
}

func (r *Router) HEAD(pattern string, handlers ...interface{}) {
	r.handle(http.MethodHead, pattern, handlers)
}

// add all route to routeStore
func dumpRoute(r *Router, rs *routeStore) {
	if r.sub == nil { // end route
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

		if len(tmp.gbefores) > 0 {
			hstmp := make([]interface{}, len(tmp.gbefores))
			copy(hstmp, tmp.gbefores)

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

func (r *Router) IsParent() bool {
	return r.parent == nil
}

// to generate router tree.
// r is root router.
func (r *Router) Handler(opts ...Option) *Engine {
	if !r.IsParent() {
		panic("sub router not allowed: Handler()")
	}

	o := &options{}
	for _, f := range opts {
		f(o)
	}

	// for global middleware
	if len(o.NoFoundHandlers) > 0 && len(r.gbefores) > 0 {
		hstmp := make([]Handler, len(r.gbefores))
		copy(hstmp, newHandlers(r.gbefores))

		hstmp = append(hstmp, o.NoFoundHandlers...)
		o.NoFoundHandlers = hstmp
	}

	rs := newRouteStore()

	dumpRoute(r, rs)

	// if len(rs.routeSlice) == 0 {
	// 	panic("no route: Handler()")
	// }

	// check uri
	for _, v := range rs.routeSlice {
		if !(v.uri == "/" || checkSplitPattern(v.uri)) {
			panic(fmt.Sprintf("invalid route : [%s : %s]", v.method, v.uri))
		}

		v.variantUri = _VariantUri(v.uri)
	}

	w := newWater()

	w.rootRouter = r
	w.routeStore = rs
	w.options = o

	defaultMultipartMemory = w.options.MaxMultipartMemory
	binding.SetMultipartMemory(defaultMultipartMemory)

	w.buildTree()

	return w
}

func _VariantUri(raw string) string {
	if !strings.Contains(raw, "/:") && !strings.Contains(raw, "/*") {
		return raw
	}

	ls := strings.Split(raw, "/")

	for i, v := range ls {
		if strings.HasPrefix(v, ":") {
			ls[i] = "<" + strings.TrimSpace(v[1:]) + ">"
		}
		if strings.HasPrefix(v, "*") {
			ls[i] = strings.TrimSpace(v)
		}
	}

	return strings.Join(ls, "/")
}

func (r *Router) Classic() {
	if !r.IsParent() {
		panic("sub router not allowed : Classic()")
	}

	r.Use(Logger())
	r.Use(Recovery())
}

// Default returns an router instance with the Logger and Recovery middleware already attached.
func Default() *Router {
	r := NewRouter()
	r.Classic()

	return r
}
