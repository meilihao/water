package water

type options struct {
	EnableStaticRouter bool
	NoFoundHandler     Handler
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

// WithNoFoundHandler the handler for no match route
// for vue spa
func WithNoFoundHandler(h Handler) Option {
	return func(o *options) {
		o.NoFoundHandler = h
	}
}

// WithMaxMultipartMemory is given to http.Request's ParseMultipartForm method call.
func WithMaxMultipartMemory(max int64) Option {
	return func(o *options) {
		o.MaxMultipartMemory = max
	}
}
