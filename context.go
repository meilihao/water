package water

import (
	"net/http"

	"github.com/meilihao/logx"
)

// Context represents the context of current request of water instance.
// Context体现了water处理当前请求时的上下文环境
// 不在Context中使用"Resp  ResponseWriter"的原因:Resp.Write()和Ctx.Write()的行为不一致,
// 因Resp.Write()不会设置ctx.written,导致在Recovery()等地方重复调用WriteHeader而报错
// "http: multiple response.WriteHeader calls".
type Context struct {
	Environ Environ
	Params  Params
	Req     *http.Request
	ResponseWriter

	handlers       []Handler
	handlersLength int
	index          int

	written bool
	status  int

	Id string
}

type ResponseWriter interface {
	http.ResponseWriter
	//http.Flusher
	//http.Hijacker
}

func newContext() *Context {
	return &Context{}
}

func (ctx *Context) reset() {
	ctx.index = 0

	ctx.written = false
	ctx.status = 0
}

func (ctx *Context) Next() {
	ctx.index += 1
	ctx.run()
}

func (ctx *Context) run() {
	for ctx.index < ctx.handlersLength {
		ctx.handlers[ctx.index].ServeHTTP(ctx)
		ctx.index += 1

		if ctx.written {
			return
		}
	}
}

func (ctx *Context) Written() bool {
	return ctx.written
}

func (ctx *Context) Status() int {
	return ctx.status
}

func (ctx *Context) WriteHeader(code int) {
	if ctx.written {
		logx.Warn("water: multiple ctx.WriteHeader calls")
	} else {
		ctx.written = true
		ctx.status = code
		ctx.ResponseWriter.WriteHeader(code)
	}
}

func (ctx *Context) Write(data []byte) (int, error) {
	if !ctx.written {
		header := ctx.Header()
		if len(data) > 0 && header.Get("Content-Type") == "" {
			header.Set("Content-Type", http.DetectContentType(data))
		}
		ctx.WriteHeader(http.StatusOK)
	}
	return ctx.ResponseWriter.Write(data)
}
