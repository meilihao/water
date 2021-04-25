package binding

import (
	"net/http"
)

const (
	MIMEJSON              = "application/json"
	MIMEXML               = "application/xml"
	MIMEXML2              = "text/xml"
	MIMEPlain             = "text/plain"
	MIMEPOSTForm          = "application/x-www-form-urlencoded"
	MIMEMultipartPOSTForm = "multipart/form-data"
)

type Bindinger interface {
	Name() string
	Bind(*http.Request, interface{}) error
}

var (
	JSON          = jsonBinding{}
	XML           = xmlBinding{}
	Form          = formBinding{}
	FormPost      = formBinding{}
	FormMultipart = formMultipartBinding{}
	Query         = queryBinding{}
	Header        = headerBinding{}
	Uri           = uriBinding{}
)

// NewBindinger returns the appropriate Binding instance based on HTTP method
// and content type.
func NewBindinger(method, contentType string) Bindinger {
	if method == http.MethodGet {
		return Form
	}

	if (method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch) &&
		contentType == MIMEMultipartPOSTForm {
		return FormMultipart
	}

	switch contentType {
	case MIMEJSON:
		return JSON
	case MIMEXML, MIMEXML2:
		return XML
	case MIMEPOSTForm:
		return Form
	default:
		return Form
	}
}

func validate(obj interface{}) error {
	if Validator == nil {
		return nil
	}
	return Validator.ValidateStruct(obj)
}
