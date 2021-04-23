package water

import (
	"net/http"
	"path"
	"path/filepath"
	"strings"
)

type ETagType byte

const (
	ETagTypeNone = iota
	ETagType1
	ETagType2
)

// StaticOptions is a struct for specifying configuration options for the Static Files middleware.
type StaticOptions struct {
	// Prefix is the optional prefix used to serve the static directory content
	Prefix string
	// IndexFile defines which file to serve as index if it exists.
	IndexFile string
	hasIndex  bool
	// Expires defines which user-defined function to use for producing a HTTP Expires Header
	Expires func() string
	// ETag defines if we should add an ETag header
	ETag ETagType
	// FileSystem is the interface for supporting any implmentation of file system.
	FileSystem http.FileSystem
}

// Static serves files from the given file system root.
// http.NotFound is used instead of the Router's NotFound handler.
// can send expires, etag, index
func (r *Router) StaticAdvance(opt *StaticOptions) {
	if r.parent != nil {
		panic("sub router not allowed : Static()")
	}

	if len(opt.IndexFile) != 0 {
		opt.hasIndex = true
	}

	if opt.Prefix != "" {
		if opt.Prefix[0] != '/' {
			opt.Prefix = "/" + opt.Prefix
		}
		opt.Prefix = strings.TrimRight(opt.Prefix, "/")
	}
	if opt.FileSystem == nil {
		panic("need FileSystem")
	}

	h := func(ctx *Context) {
		f, err := opt.FileSystem.Open(ctx.Param("*0"))
		if err != nil {
			ctx.WriteHeader(http.StatusNotFound)

			return
		}
		defer f.Close()

		fi, err := f.Stat()
		if err != nil {
			ctx.WriteHeader(http.StatusNotFound) // File exists but fail to open.

			return
		}

		if opt.Expires != nil {
			ctx.Header().Set("Expires", opt.Expires())
		}

		if fi.IsDir() {
			redirPath := path.Clean(ctx.Request.URL.Path)
			// path.Clean removes the trailing slash, so we need to add it back when
			// the original path has it.
			if strings.HasSuffix(ctx.Request.URL.Path, "/") {
				redirPath = redirPath + "/"
			}
			// Redirect if missing trailing slash.
			if !strings.HasSuffix(redirPath, "/") {
				http.Redirect(ctx.ResponseWriter, ctx.Request, redirPath+"/", http.StatusFound)
				return
			}

			ctx.File(filepath.Join(ctx.Param("*0"), opt.IndexFile))
			return
		}

		if opt.ETag != ETagTypeNone {
			var tag string

			if opt.ETag == ETagType1 {
				tag = GenerateETag(fi.ModTime(), fi.Size())
			} else {
				tag = GenerateETag2(fi.ModTime(), fi.Size(), fi.Name())
			}

			ctx.Header().Set("ETag", tag)
			if ctx.Request.Header.Get("If-None-Match") == tag {
				ctx.WriteHeader(http.StatusNotModified)
				return
			}
		}

		http.ServeContent(ctx.ResponseWriter, ctx.Request, ctx.Param("*0"), fi.ModTime(), f)
	}

	r.GET(filepath.Join(opt.Prefix, "/*"), h)
	r.HEAD(filepath.Join(opt.Prefix, "/*"), h)
}

// Static serves files from the given file system root.
// http.NotFound is used instead of the Router's NotFound handler.
// use :
//     router.Static("/static", "/var/www")
func (r *Router) Static(uri, root string) {
	opt := &StaticOptions{
		Prefix:     uri,
		FileSystem: http.Dir(root),
	}

	r.StaticAdvance(opt)
}

// StaticFile registers a single route in order to serve a single file
// router.StaticFile("favicon.ico", "./resources/favicon.ico")
func (r *Router) StaticFile(relativePath, filepath string) {
	if r.parent != nil {
		panic("sub router not allowed : Static()")
	}

	handler := func(c *Context) {
		c.File(filepath)
	}

	r.GET(relativePath, handler)
	r.HEAD(relativePath, handler)
}
