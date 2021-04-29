package water

type options struct {
	// 适用于多静态路由的场景
	EnableStaticRouter bool
	NoFoundHandler     Handler
}

type Option func(*options)

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
