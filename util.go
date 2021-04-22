package water

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"reflect"
	"runtime"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
)

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary
)

// from net/http
func cleanPath(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}

	// Restful uri
	// path.Clean会去除末尾的'/'("/"除外)
	return path.Clean(p)
}

func requestProxy(req *http.Request) []string {
	if ips := req.Header.Get(HeaderXForwardedFor); ips != "" {
		return strings.Split(ips, ", ")
	}

	return nil
}

// see more detail in github.com/minio/minio/pkg/handlers/proxy.go
func requestRealIp(req *http.Request) string {
	ip := req.Header.Get(HeaderXRealIP)
	if ip == "" {
		ips := requestProxy(req)
		if len(ips) > 0 && ips[0] != "" {
			return ips[0]
		}

		ip = req.RemoteAddr
		if i := strings.LastIndex(ip, ":"); i > -1 {
			ip = ip[:i]
		}
	}

	return ip
}

func contentDisposition(fileName, dispositionType string) string {
	if dispositionType == "" {
		dispositionType = "attachment"
	}
	if fileName == "" {
		return dispositionType
	}

	return fmt.Sprintf(`%s; filename="%s"; filename*=UTF-8''%s`,
		dispositionType, url.PathEscape(fileName), url.PathEscape(fileName))
}

// check pattern segment
// 检查url片段的合法性
func checkSplitPattern(pattern string) bool {
	n := len(pattern)
	return n > 0 && pattern[0] == '/' && pattern[n-1] != '/'
}

func reverseStrings(s []string) []string {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}

	return s
}

func checkMethod(method string) (string, int) {
	method = strings.ToUpper(method)
	idx := MethodIndex(method)
	if idx < 0 {
		panic("unsupport method: " + method)
	}

	return method, idx
}

func filterFlags(content string) string {
	for i, char := range content {
		if char == ' ' || char == ';' {
			return content[:i]
		}
	}
	return content
}

func nameOfFunction(f interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}

// GenerateETag generates an ETag based on file modification time and size
// from http://lxr.nginx.org/ident?_i=ngx_http_set_etag
func GenerateETag(lastModified time.Time, size int64) string {
	return fmt.Sprintf("%x-%x", lastModified.Unix(), size)
}

func resolveAddress(addr []string) string {
	switch len(addr) {
	case 0:
		return ":8080"
	case 1:
		return addr[0]
	default:
		panic("too many addrs")
	}
}
