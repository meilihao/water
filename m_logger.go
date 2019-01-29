package water

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/meilihao/logx"
)

var (
	green  = string([]byte{27, 91, 57, 55, 59, 52, 50, 109})
	white  = string([]byte{27, 91, 57, 48, 59, 52, 55, 109})
	yellow = string([]byte{27, 91, 57, 55, 59, 52, 51, 109})
	red    = string([]byte{27, 91, 57, 55, 59, 52, 49, 109})
	reset  = string([]byte{27, 91, 48, 109})

	LogColor      = true
	LogClose      = false
	LogHttpBody   = true
	LogTimeFormat = "2006-01-02 15:04:05"
)

func colorForStatus(code int) string {
	switch {
	case code >= 200 && code < 300:
		return green
	case code >= 300 && code < 400:
		return white
	case code >= 400 && code < 500:
		return yellow
	default:
		return red
	}
}

func logStatus(status int) string {
	if LogColor {
		return fmt.Sprintf("%s %3d %s", colorForStatus(status), status, reset)
	} else {
		return fmt.Sprintf(" %3d ", status)
	}
}

func logPrefix(req *http.Request) string {
	return "[ water : " + req.Header.Get("Trace-ID") + " ]"
}

func Logger() HandlerFunc {
	return func(ctx *Context) {
		if LogClose {
			return
		}

		start := time.Now()
		body := ""

		if LogHttpBody && ctx.Request.Body != nil && strings.Contains(ctx.Request.Header.Get("Content-Type"), "application/json") {
			safe := &io.LimitedReader{R: ctx.Request.Body, N: 1 << 20}
			requestbody, _ := ioutil.ReadAll(safe)
			ctx.Request.Body.Close()

			bf := bytes.NewBuffer(requestbody)
			ctx.Request.Body = ioutil.NopCloser(bf)

			body = "\n" + string(requestbody)
		}

		ctx.Next()

		// Layout : "prefix start_time [ status ] used_time | ip | method path"
		logx.Infof("%s %v |%s| %13v | %16s | %7s %s%s",
			logPrefix(ctx.Request),
			start.Format(LogTimeFormat),
			logStatus(ctx.status),
			time.Now().Sub(start),
			ctx.RealIp(),
			ctx.Request.Method,
			ctx.Request.URL.String(),
			body,
		)
	}
}
