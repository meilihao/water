package water

type options struct {
	EnableStaticRouter bool
	NoFoundHandler     Handler
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
