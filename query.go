package water

import (
	"strconv"
	"strings"
)

func (ctx *Context) queryExist(name string) bool {
	if ctx.Request.Form == nil {
		return false
	}
	if len(ctx.Request.Form[name]) == 0 {
		return false
	}

	return true
}

// Query returns trimmed string.
// Note: It is recommended! If not, you can use "ctx.Request.FormValue".
// 这是推荐的做法,如果不认同,可使用ctx.Request.FormValue.
func (ctx *Context) Query(name string) string {
	ctx.ParseFormOrMultipartForm()

	if !ctx.queryExist(name) {
		return ""
	}

	return strings.TrimSpace(ctx.Request.Form.Get(name))
}

// QueryBool returns bool
func (ctx *Context) QueryBool(name string) bool {
	v, _ := strconv.ParseBool(ctx.Query(name))
	return v
}

// QueryInt returns int
func (ctx *Context) QueryInt(name string) int {
	v, _ := strconv.Atoi(ctx.Query(name))
	return v
}

// QueryInt64 returns int64
func (ctx *Context) QueryInt64(name string) int64 {
	v, _ := strconv.ParseInt(ctx.Query(name), 10, 64)
	return v
}

// QueryUint returns uint
func (ctx *Context) QueryUint(name string) uint {
	v, _ := strconv.ParseUint(ctx.Query(name), 10, 64)
	return uint(v)
}

// QueryUint64 returns uint64
func (ctx *Context) QueryUint64(name string) uint64 {
	v, _ := strconv.ParseUint(ctx.Query(name), 10, 64)
	return v
}

// QueryFloat64 returns float64
func (ctx *Context) QueryFloat64(name string) float64 {
	v, _ := strconv.ParseFloat(ctx.Query(name), 64)
	return v
}

// QueryArray returns a list of results by given query name
func (ctx *Context) QueryArray(name string) []string {
	ctx.ParseFormOrMultipartForm()

	if !ctx.queryExist(name) {
		return nil
	}

	vs := ctx.Request.Form[name]

	tmp := make([]string, len(vs))
	for i := range vs {
		tmp[i] = strings.TrimSpace(vs[i])
	}
	return tmp
}

// QueryMap returns a map for a given query key.
// POST /post?ids[a]=1234&ids[b]=hello HTTP/1.1
// Content-Type: application/x-www-form-urlencoded
// names[first]=thinkerou&names[second]=tianou
func (ctx *Context) QueryMap(key string) map[string]string {
	ctx.ParseFormOrMultipartForm()

	dicts := make(map[string]string)

	for k, v := range ctx.Request.Form {
		if i := strings.IndexByte(k, '['); i >= 1 && k[0:i] == key {
			if j := strings.IndexByte(k[i+1:], ']'); j >= 1 {
				dicts[k[i+1:][:j]] = strings.TrimSpace(v[0])
			}
		}
	}

	return dicts
}

// DefaultQuery returns trimmed string, but no name return default
// Note: It is recommended! If not, you can use "ctx.Request.FormValue".
// 这是推荐的做法,如果不认同,可使用ctx.Request.FormValue.
func (ctx *Context) DefaultQuery(name, defaultValue string) string {
	ctx.ParseFormOrMultipartForm()

	if !ctx.queryExist(name) {
		return defaultValue
	}

	return strings.TrimSpace(ctx.Request.Form.Get(name))
}

// DefaultQueryBool returns bool
func (ctx *Context) DefaultQueryBool(name, defaultValue string) bool {
	v, _ := strconv.ParseBool(ctx.DefaultQuery(name, defaultValue))
	return v
}

// DefaultQueryInt returns int
func (ctx *Context) DefaultQueryInt(name, defaultValue string) int {
	v, _ := strconv.Atoi(ctx.DefaultQuery(name, defaultValue))
	return v
}

// DefaultQueryInt64 returns int64
func (ctx *Context) DefaultQueryInt64(name, defaultValue string) int64 {
	v, _ := strconv.ParseInt(ctx.DefaultQuery(name, defaultValue), 10, 64)
	return v
}

// DefaultQueryUint returns uint
func (ctx *Context) DefaultQueryUint(name, defaultValue string) uint {
	v, _ := strconv.ParseUint(ctx.DefaultQuery(name, defaultValue), 10, 64)
	return uint(v)
}

// DefaultQueryUint64 returns uint64
func (ctx *Context) DefaultQueryUint64(name, defaultValue string) uint64 {
	v, _ := strconv.ParseUint(ctx.DefaultQuery(name, defaultValue), 10, 64)
	return v
}

// DefaultQueryFloat64 returns float64
func (ctx *Context) DefaultQueryFloat64(name, defaultValue string) float64 {
	v, _ := strconv.ParseFloat(ctx.DefaultQuery(name, defaultValue), 64)
	return v
}

// DefaultQueryArray returns a list of results by given query name
func (ctx *Context) DefaultQueryArray(name string, defaultValue []string) []string {
	ctx.ParseFormOrMultipartForm()

	vs, ok := ctx.Request.Form[name]
	if !ok {
		return defaultValue
	}

	tmp := make([]string, len(vs))
	for i := range vs {
		tmp[i] = strings.TrimSpace(vs[i])
	}
	return tmp
}

// DefaultQueryMap returns a map for a given query key.
func (ctx *Context) DefaultQueryMap(key string, defaultValue map[string]string) map[string]string {
	ctx.ParseFormOrMultipartForm()

	dicts := make(map[string]string)
	exist := false

	for k, v := range ctx.Request.Form {
		if i := strings.IndexByte(k, '['); i >= 1 && k[0:i] == key {
			if j := strings.IndexByte(k[i+1:], ']'); j >= 1 {
				exist = true

				dicts[k[i+1:][:j]] = strings.TrimSpace(v[0])
			}
		}
	}

	if !exist {
		return defaultValue
	}

	return dicts
}
