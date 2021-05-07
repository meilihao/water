package water

const (
	DEFAULT_TPL_SET_NAME = ""
)

var _render Render

func SetRender(r Render) {
	_render = r
}

type defaultRender struct{}

func (r *defaultRender) HTMLSet(ctx *Context, code int, setName, tplName string, data interface{}) {
	ctx.WriteHeader(code)
	ctx.Header().Set(HeaderContentType, MIMETextHTMLCharsetUTF8)
}

func init() {
	_render = new(defaultRender)
}

type Render interface {
	HTMLSet(*Context, int, string, string, interface{})
}

// HTML renders the HTML with default template set.
func (ctx *Context) HTML(code int, name string, data interface{}) {
	ctx.renderHTML(code, DEFAULT_TPL_SET_NAME, name, data)
}

// HTMLSet renders the HTML with given template set name.
func (ctx *Context) HTMLSet(code int, setName, tplName string, data interface{}) {
	ctx.renderHTML(code, setName, tplName, data)
}

func (ctx *Context) renderHTML(code int, setName, tplName string, data interface{}) {
	_render.HTMLSet(ctx, code, setName, tplName, data)
}
