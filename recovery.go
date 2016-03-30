package water

import (
	"fmt"
	"net/http"
	"runtime"
	"time"
)

func Recovery() HandlerFunc {
	return func(ctx *Context) {
		defer func() {
			if err := recover(); err != nil {
				content := fmt.Sprintf("HandlerFunc crashed with error: %v", err)
				for skip := 1; ; skip += 1 {
					//当前goroutine的栈上的函数调用信息，主要有当前的PC值和被调用的文件及其行号.
					//skip==0表示当前栈,这里是指Recovery().
					_, file, line, ok := runtime.Caller(skip)
					if !ok {
						break
					} else {
						content += "\n"
					}
					content += fmt.Sprintf("%v %v", file, line)
				}

				if !ctx.written {
					ctx.WriteHeader(http.StatusInternalServerError)
				}

				ctx.router.logger.Printf("panic %s : %s\n", time.Now().Format(LogTimeFormat), content)

				if Status != Stable {
					ctx.WriteString(content)
				}
			} else {
				if !ctx.written {
					ctx.WriteHeader(http.StatusOK)
				}
			}
		}()

		ctx.Next()
	}
}
