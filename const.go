package water

const (
	MIMESeparator = "; "
	CharsetUTF8   = "charset=utf-8"
)

// MIME types
const (
	MIMEApplicationJSON            = "application/json"
	MIMEApplicationJSONCharsetUTF8 = MIMEApplicationJSON + MIMESeparator + CharsetUTF8
	MIMEApplicationXML             = "application/xml"
	MIMEApplicationXMLCharsetUTF8  = MIMEApplicationXML + MIMESeparator + CharsetUTF8
	MIMETextHTML                   = "text/html"
	MIMETextHTMLCharsetUTF8        = MIMETextHTML + MIMESeparator + CharsetUTF8
	MIMETextPlain                  = "text/plain"
	MIMETextPlainCharsetUTF8       = MIMETextPlain + MIMESeparator + CharsetUTF8
	MIMEXMLHttpRequest             = "XMLHttpRequest"
)

// HTTP Header Fields, from chrome
const (
	HeaderCacheControl    = "Cache-Control"
	HeaderContentType     = "Content-Type"
	HeaderExpires         = "Expires"
	HeaderUserAgent       = "User-Agent"
	HeaderXForwardedFor   = "X-Forwarded-For"
	HeaderXForwardedProto = "X-Forwarded-Proto"
	HeaderXRealIP         = "X-Real-IP"
	HeaderXRequestedWith  = "X-Requested-With"
)
