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
// see https://github.com/teambition/gear/blob/master/const.go
const (
	HeaderCacheControl = "Cache-Control" // Requests, Responses
	HeaderContentType  = "Content-Type"  // Requests, Responses

	HeaderUserAgent      = "User-Agent"       // Requests
	HeaderXRequestedWith = "X-Requested-With" // Requests

	HeaderExpires            = "Expires"             // Responses
	HeaderContentDisposition = "Content-Disposition" // Responses

	// Common Non-Standard Response Headers
	HeaderXForwardedFor   = "X-Forwarded-For"   // Requests
	HeaderXForwardedProto = "X-Forwarded-Proto" // Requests
	HeaderXRealIP         = "X-Real-IP"         // Requests
)
