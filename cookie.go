package water

import (
	"encoding/hex"
	"html/template"
	"net/http"
	"strconv"
	"time"
)

func (ctx *Context) Cookie(name string) string {
	cookie, err := ctx.Req.Cookie(name)
	if err != nil {
		return ""
	}

	return cookie.Value
}

func (ctx *Context) CookieEscape(name string) string {
	return template.HTMLEscapeString(ctx.Cookie(name))
}

func (ctx *Context) CookieBool(name string) bool {
	v, _ := strconv.ParseBool(ctx.Cookie(name))
	return v
}

func (ctx *Context) CookieInt(name string) int {
	v, _ := strconv.Atoi(ctx.Cookie(name))
	return v
}

func (ctx *Context) CookieInt64(name string) int64 {
	v, _ := strconv.ParseInt(ctx.Cookie(name), 10, 64)
	return v
}

func (ctx *Context) CookieUint(name string) uint {
	v, _ := strconv.ParseUint(ctx.Cookie(name), 10, 64)
	return uint(v)
}

func (ctx *Context) CookieUint64(name string) uint64 {
	v, _ := strconv.ParseUint(ctx.Cookie(name), 10, 64)
	return v
}

func (ctx *Context) CookieFloat64(name string) float64 {
	v, _ := strconv.ParseFloat(ctx.Cookie(name), 64)
	return v
}

func (ctx *Context) SetCookie(name string, value string, others ...interface{}) {
	cookie := http.Cookie{}
	cookie.Name = name
	cookie.Value = value

	if len(others) > 0 {
		switch v := others[0].(type) {
		case int:
			cookie.MaxAge = v
		case int32:
			cookie.MaxAge = int(v)
		case int64:
			cookie.MaxAge = int(v)
		}

		// for ie
		if cookie.MaxAge > 0 {
			cookie.Expires = time.Now().Add(time.Duration(cookie.MaxAge) * time.Second)
		}
	}

	cookie.Path = "/"
	if len(others) > 1 {
		if v, ok := others[1].(string); ok && len(v) > 0 {
			cookie.Path = v
		}
	}

	if len(others) > 2 {
		if v, ok := others[2].(string); ok && len(v) > 0 {
			cookie.Domain = v
		}
	}

	if len(others) > 3 {
		switch v := others[3].(type) {
		case bool:
			cookie.Secure = v
		default:
			if others[3] != nil {
				cookie.Secure = true
			}
		}
	}

	if len(others) > 4 {
		if v, ok := others[4].(bool); ok && v {
			cookie.HttpOnly = true
		}
	}

	ctx.ResponseWriter.Header().Add("Set-Cookie", cookie.String())
}

func (ctx *Context) SecureCookie(secret, key string) (string, bool) {
	val := ctx.Cookie(key)
	if val == "" {
		return "", false
	}

	data, err := hex.DecodeString(val)
	if err != nil {
		return "", false
	}

	text, err := AESDecrypt([]byte(secret), data)
	return string(text), err == nil
}

// http2 is Recommended
// 推荐使用http2
func (ctx *Context) SetSecureCookie(secret, name, value string, others ...interface{}) {
	text, err := AESEncrypt([]byte(secret), []byte(value))
	if err != nil {
		panic("error encrypting cookie: " + err.Error())
	}
	ctx.SetCookie(name, hex.EncodeToString(text), others...)
}
