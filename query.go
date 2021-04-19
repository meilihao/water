package water

import (
	"errors"
	"html/template"
	"strconv"
	"strings"
)

const (
	// MaxMultipartMemory
	DefaultMaxMemory = 10 << 20 // 10MB
)

// parseFormOrMultipartForm parses the raw query from the URL.
func (ctx *Context) parseFormOrMultipartForm() error {
	if (ctx.Request.Method == "POST" || ctx.Request.Method == "PUT") && strings.Contains(ctx.Request.Header.Get("Content-Type"), "multipart/form-data") {
		if err := ctx.Request.ParseMultipartForm(DefaultMaxMemory); err != nil {
			return errors.New("parseMultipartForm error:" + err.Error())
		}
	} else {
		if err := ctx.Request.ParseForm(); err != nil {
			return errors.New("parseForm error:" + err.Error())
		}
	}
	return nil
}

// QueryString returns escapred and trimmed string.
// Note: It is recommended! If not, you can use "ctx.Request.FormValue".
// 这是推荐的做法,如果不认同,可使用ctx.Request.FormValue.
func (ctx *Context) QueryString(name string) string {
	if err := ctx.parseFormOrMultipartForm(); err != nil {
		panic(err.Error())
	}
	return template.HTMLEscapeString(strings.TrimSpace(ctx.Request.Form.Get(name)))
}

// QueryBool returns bool
func (ctx *Context) QueryBool(name string) bool {
	v, _ := strconv.ParseBool(ctx.QueryString(name))
	return v
}

// QueryInt returns int
func (ctx *Context) QueryInt(name string) int {
	v, _ := strconv.Atoi(ctx.QueryString(name))
	return v
}

// QueryInt returns int64
func (ctx *Context) QueryInt64(name string) int64 {
	v, _ := strconv.ParseInt(ctx.QueryString(name), 10, 64)
	return v
}

// QueryUint returns uint
func (ctx *Context) QueryUint(name string) uint {
	v, _ := strconv.ParseUint(ctx.QueryString(name), 10, 64)
	return uint(v)
}

// QueryUint64 returns uint64
func (ctx *Context) QueryUint64(name string) uint64 {
	v, _ := strconv.ParseUint(ctx.QueryString(name), 10, 64)
	return v
}

// QueryFloat64 returns float64
func (ctx *Context) QueryFloat64(name string) float64 {
	v, _ := strconv.ParseFloat(ctx.QueryString(name), 64)
	return v
}

// QueryStrings returns a list of results by given query name
func (ctx *Context) QueryStrings(name string) []string {
	if err := ctx.parseFormOrMultipartForm(); err != nil {
		panic(err.Error())
	}

	vs, ok := ctx.Request.Form[name]
	if !ok {
		return nil
	}

	tmp := make([]string, len(vs))
	for i := range vs {
		tmp[i] = template.HTMLEscapeString(strings.TrimSpace(vs[i]))
	}
	return tmp
}
