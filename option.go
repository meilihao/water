package water

type options struct {
	// 适用于多静态路由的场景
	EnableStaticRouter bool
}

type Option func(*options)

func WithStaticRouter(enable bool) Option {
	return func(o *options) {
		o.EnableStaticRouter = enable
	}
}
