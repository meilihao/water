package water

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

// 301
func (ctx *Context) MovedPermanently(uri string) {
	http.Redirect(ctx, ctx.Req, uri, http.StatusMovedPermanently)
}

// 302
func (ctx *Context) Found(uri string) {
	http.Redirect(ctx, ctx.Req, uri, http.StatusFound)
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
	header.Set("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1.
	//header.Header().Set("Pragma", "no-cache")                          // HTTP 1.0.
	header.Set("Expires", "0") // Proxies.
}

func (ctx *Context) UserAgent() string {
	return ctx.Req.Header.Get("User-Agent")
}

// IsAjax returns boolean of this request is generated by ajax.
func (ctx *Context) IsAjax() bool {
	return ctx.Req.Header.Get("X-Requested-With") == "XMLHttpRequest"
}

// IP returns request client ip.
// if using proxy, return first proxy ip.
func (ctx *Context) RemoteIp() string {
	return requestRemoteIp(ctx.Req)
}

// Proxy returns slice of proxy client ips.
func (ctx *Context) Proxy() []string {
	return requestProxy(ctx.Req)
}

// BodyString returns content of request body in string.
func (ctx *Context) BodyString() (string, error) {
	data, err := ctx.BodyBytes()
	return string(data), err
}

// BodyBytes returns content of request body in bytes.
func (ctx *Context) BodyBytes() ([]byte, error) {
	data, err := ioutil.ReadAll(ctx.Req.Body)
	if err != nil {
		return nil, err
	}
	ctx.Req.Body.Close()
	return data, err
}

func (ctx *Context) FormFile(key string) (multipart.File, *multipart.FileHeader, error) {
	return ctx.Req.FormFile(key)
}

func (ctx *Context) SaveToFile(fileName, savePath string) error {
	file, _, err := ctx.Req.FormFile(fileName)
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
	http.ServeFile(ctx, ctx.Req, filePath)
}

func (ctx *Context) Download(fpath string) error {
	f, err := os.Open(fpath)
	if err != nil {
		return err
	}
	defer f.Close()

	fName := filepath.Base(fpath)
	ctx.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%v\"", fName))
	_, err = io.Copy(ctx, f)
	return err
}

func (ctx *Context) WriteString(str string) {
	ctx.Header().Set("Content-Type", "text/plain; charset=utf-8")
	ctx.Write([]byte(str))
}

func (ctx *Context) WriteJson(v interface{}) error {
	ctx.Header().Set("Content-Type", "application/json; charset=UTF-8")
	err := json.NewEncoder(ctx).Encode(v)
	if err != nil {
		ctx.Header().Del("Content-Type")
	}
	return err
}

func (ctx *Context) WriteXml(v interface{}) error {
	ctx.Header().Set("Content-Type", "application/xml; charset=UTF-8")
	err := xml.NewEncoder(ctx).Encode(v)
	if err != nil {
		ctx.Header().Del("Content-Type")
	}
	return err
}

func (ctx *Context) DecodeJson(v interface{}) error {
	defer ctx.Req.Body.Close()

	return json.NewDecoder(ctx.Req.Body).Decode(v)
}

func (ctx *Context) DecodeXml(v interface{}) error {
	defer ctx.Req.Body.Close()

	return xml.NewDecoder(ctx.Req.Body).Decode(v)
}

func (ctx *Context) ErrorJson(v string) {
	ctx.WriteJson(map[string]string{"Error": v})
}

func (ctx *Context) IdJson(v interface{}) {
	ctx.WriteJson(map[string]interface{}{"Id": v})
}
