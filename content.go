package water

import (
	"encoding/xml"
	"errors"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/meilihao/logx"
)

type H map[string]interface{}

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
	ctx.Header().Set(HeaderCacheControl, "no-cache, max-age=0, s-max-age=0, must-revalidate") // HTTP 1.1
	ctx.Header().Set("Expires", "0")                                                          // Proxies.
}

func (ctx *Context) UserAgent() string {
	return ctx.Request.Header.Get(HeaderUserAgent)
}

// IsAjax returns boolean of this request is generated by ajax.
func (ctx *Context) IsAjax() bool {
	return ctx.Request.Header.Get(HeaderXRequestedWith) == MIMEXMLHttpRequest
}

func (ctx *Context) IsWebsocket() bool {
	if strings.Contains(strings.ToLower(ctx.Request.Header.Get("Connection")), "upgrade") &&
		strings.ToLower(ctx.Request.Header.Get("Upgrade")) == "websocket" {
		return true
	}

	return false
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

// ContentType returns the Content-Type header of the request.
func (ctx *Context) ContentType() string {
	return filterFlags(ctx.Request.Header.Get("Content-Type"))
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

// ReadCloser returns a ReadCloser for request body, need Close()
func (ctx *Context) ReadCloser() io.ReadCloser {
	return ctx.Request.Body
}

// File writes the specified file into the body stream.
func (ctx *Context) File(filepath string) {
	http.ServeFile(ctx.ResponseWriter, ctx.Request, filepath)
}

func (ctx *Context) FormFile(name string) (multipart.File, *multipart.FileHeader, error) {
	return ctx.Request.FormFile(name)
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

func (ctx *Context) Download(fpath string, inline ...bool) error {
	f, err := os.Open(fpath)
	if err != nil {
		return err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return err
	}

	if fi.IsDir() {
		return errors.New("This is Dir.")
	}

	dispositionType := "attachment"
	if len(inline) > 0 && inline[0] {
		dispositionType = "inline"
	}

	ctx.Header().Set("Content-Description", "File Transfer")
	ctx.Header().Set("Content-Type", "application/octet-stream")
	ctx.Header().Set("Content-Transfer-Encoding", "binary")
	ctx.Header().Set("Expires", "0")
	ctx.Header().Set("Cache-Control", "must-revalidate")
	ctx.Header().Set("Pragma", "public")
	ctx.Header().Set(HeaderContentDisposition, contentDisposition(fi.Name(), dispositionType))
	http.ServeContent(ctx, ctx.Request, fi.Name(), fi.ModTime(), f)

	return nil
}

func (ctx *Context) Attachment(name string, modtime time.Time, content io.ReadSeeker, inline ...bool) {
	dispositionType := "attachment"
	if len(inline) > 0 && inline[0] {
		dispositionType = "inline"
	}

	ctx.Header().Set(HeaderContentDisposition, contentDisposition(name, dispositionType))
	http.ServeContent(ctx, ctx.Request, name, modtime, content)
}

func (ctx *Context) Stream(contentType string, r io.Reader) error {
	ctx.Header().Set(HeaderContentType, contentType)

	_, err := io.Copy(ctx, r)

	return err
}

func (ctx *Context) String(code int, str string) {
	ctx.WriteHeader(code)
	ctx.Header().Set(HeaderContentType, MIMETextPlainCharsetUTF8)
	ctx.Write([]byte(str))
}

func (ctx *Context) HTML(code int, str string) {
	ctx.WriteHeader(code)
	ctx.Header().Set(HeaderContentType, MIMETextHTMLCharsetUTF8)
	ctx.Write([]byte(str))
}

func (ctx *Context) JSON(code int, v interface{}) error {
	ctx.WriteHeader(code)
	ctx.Header().Set(HeaderContentType, MIMEApplicationJSONCharsetUTF8)
	err := json.NewEncoder(ctx).Encode(v)
	if err != nil {
		logx.Warn(err)
	}
	return err
}

func (ctx *Context) XML(code int, v interface{}) error {
	ctx.WriteHeader(code)
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

// HandlerName returns the last handler's name.
// For example if the handler is "_Users()", this function will return "main._Users".
func (ctx *Context) HandlerName() string {
	return nameOfFunction(ctx.handlers[ctx.handlersLength-1])
}

// `DefaultNotFoundHandler.FilterPath=func(c){return "/index.html"}`用于spa时可能会无限重定向, 原因是http.FileServer对路径"/index.html"有特别处理
type DefaultNotFoundHandler struct {
	FileServer http.Handler
	FilterPath func(string) string
}

func NewDefaultNotFoundHandler(root string) *DefaultNotFoundHandler {
	return &DefaultNotFoundHandler{
		FileServer: http.FileServer(http.Dir(root)),
	}
}

func (h *DefaultNotFoundHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.FilterPath != nil {
		r.URL.Path = h.FilterPath(r.URL.Path)
	}
	h.FileServer.ServeHTTP(w, r)
}

type SPANotFoundHandler struct {
	Index string
}

func NewSPANotFoundHandler(index string) *SPANotFoundHandler {
	return &SPANotFoundHandler{
		Index: index,
	}
}

func (h *SPANotFoundHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, h.Index)
}
