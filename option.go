package water

type options struct {
	EnableStaticRouter bool
	NoFoundHandlers    []Handler
	MaxMultipartMemory int64
}

type Option func(*options)

// WithStaticRouter for the scene of multi status route
// 适用于多静态路由的场景
func WithStaticRouter(enable bool) Option {
	return func(o *options) {
		o.EnableStaticRouter = enable
	}
}

// WithNoFoundHandlers the handler for no match route, example: vue spa
// code=404, can use middleware
func WithNoFoundHandlers(hs ...interface{}) Option {
	if len(hs) == 0 {
		panic("no NoFoundHandlers")
	}

	return func(o *options) {
		o.NoFoundHandlers = newHandlers(hs)
	}
}

// WithMaxMultipartMemory is given to http.Request's ParseMultipartForm method call.
func WithMaxMultipartMemory(max int64) Option {
	return func(o *options) {
		o.MaxMultipartMemory = max
	}
}
