package water

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/meilihao/logx"
)

func (ctx *Context) GetHeader(key string) string {
	return ctx.Request.Header.Get(key)
}

func (ctx *Context) SetHeader(key, value string) {
	ctx.Header().Set(key, value)
}

// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Forwarded-Proto
func (ctx *Context) Protocol() string {
	switch s := ctx.GetHeader(HeaderXForwardedProto); s {
	case "http", "https":
		return s
	}

	if ctx.Request.TLS != nil {
		return "https"
	}

	return "http"
}

// 301
func (ctx *Context) MovedPermanently(uri string) {
	http.Redirect(ctx, ctx.Request, uri, http.StatusMovedPermanently)
}

// 302
func (ctx *Context) Found(uri string) {
	http.Redirect(ctx, ctx.Request, uri, http.StatusFound)
}

// 304
func (ctx *Context) NotModified() {
	ctx.Abort(http.StatusNotModified)
}

// 400
func (ctx *Context) BadRequest() {
	ctx.Abort(http.StatusBadRequest)
}

// 401
func (ctx *Context) Unauthorized() {
	ctx.Abort(http.StatusUnauthorized)
}

// 403
func (ctx *Context) Forbidden() {
	ctx.Abort(http.StatusForbidden)
}

// 404
func (ctx *Context) NotFound() {
	ctx.Abort(http.StatusNotFound)
}

// 500
func (ctx *Context) InternalServerError() {
	ctx.Abort(http.StatusInternalServerError)
}

func (ctx *Context) Abort(code int) {
	ctx.WriteHeader(code)
}

// http://stackoverflow.com/questions/49547/making-sure-a-web-page-is-not-cached-across-all-browsers
func (ctx *Context) NoCache() {
	header := ctx.Header()
	header.Set(HeaderCacheControl, "no-cache, max-age=0, s-max-age=0, must-revalidate") // HTTP 1.1
	header.Set("Expires", "0")                                                          // Proxies.
}

func (ctx *Context) UserAgent() string {
	return ctx.Request.Header.Get(HeaderUserAgent)
}

// IsAjax returns boolean of this request is generated by ajax.
func (ctx *Context) IsAjax() bool {
	return ctx.Request.Header.Get(HeaderXRequestedWith) == MIMEXMLHttpRequest
}

// IP returns request client ip.
// if using proxy, return first proxy ip.
func (ctx *Context) RealIp() string {
	return requestRealIp(ctx.Request)
}

// Proxy returns slice of proxy client ips.
func (ctx *Context) Proxy() []string {
	return requestProxy(ctx.Request)
}

// BodyString returns content of request body in string.
func (ctx *Context) BodyString() (string, error) {
	data, err := ctx.BodyBytes()
	return string(data), err
}

// BodyBytes returns content of request body in bytes.
func (ctx *Context) BodyBytes() ([]byte, error) {
	data, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		return nil, err
	}
	ctx.Request.Body.Close()
	return data, err
}

func (ctx *Context) FormFile(key string) (multipart.File, *multipart.FileHeader, error) {
	return ctx.Request.FormFile(key)
}

func (ctx *Context) SaveToFile(fileName, savePath string) error {
	file, _, err := ctx.Request.FormFile(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	f, err := os.OpenFile(savePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, file)
	return err
}

func (ctx *Context) ServeFile(filePath string) {
	http.ServeFile(ctx, ctx.Request, filePath)
}

func (ctx *Context) Download(fpath string) error {
	f, err := os.Open(fpath)
	if err != nil {
		return err
	}
	defer f.Close()

	ctx.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"",
		url.QueryEscape(filepath.Base(fpath))))

	_, err = io.Copy(ctx, f)
	return err
}

func (ctx *Context) Stream(contentType string, r io.Reader) error {
	ctx.Header().Set(HeaderContentType, contentType)

	_, err := io.Copy(ctx, r)

	return err
}

func (ctx *Context) WriteString(str string) {
	ctx.Header().Set(HeaderContentType, MIMETextPlainCharsetUTF8)
	ctx.Write([]byte(str))
}

func (ctx *Context) WriteHTML(str string) {
	ctx.Header().Set(HeaderContentType, MIMETextHTMLCharsetUTF8)
	ctx.Write([]byte(str))
}

func (ctx *Context) WriteJson(v interface{}) error {
	ctx.Header().Set(HeaderContentType, MIMEApplicationJSONCharsetUTF8)
	err := json.NewEncoder(ctx).Encode(v)
	if err != nil {
		logx.Warn(err)
	}
	return err
}

func (ctx *Context) WriteXml(v interface{}) error {
	ctx.Header().Set(HeaderContentType, MIMEApplicationXMLCharsetUTF8)
	err := xml.NewEncoder(ctx).Encode(v)
	if err != nil {
		logx.Warn(err)
	}
	return err
}

func (ctx *Context) DecodeJson(v interface{}) error {
	defer ctx.Request.Body.Close()

	return json.NewDecoder(ctx.Request.Body).Decode(v)
}

func (ctx *Context) DecodeXml(v interface{}) error {
	defer ctx.Request.Body.Close()

	return xml.NewDecoder(ctx.Request.Body).Decode(v)
}

func (ctx *Context) ErrorJson(v interface{}) {
	ctx.WriteJson(map[string]string{"error": fmt.Sprint(v)})
}

func (ctx *Context) ErrorfJson(format string, a ...interface{}) {
	ctx.WriteJson(map[string]string{"error": fmt.Sprintf(format, a...)})
}

func (ctx *Context) IdJson(v interface{}) {
	ctx.WriteJson(map[string]interface{}{"id": v})
}

func (ctx *Context) DataJson(v interface{}) {
	ctx.WriteJson(map[string]interface{}{"data": v})
}
