package binding

import (
	"mime/multipart"
	"net/http"
	"reflect"
)

const (
	defaultMemory = 32 << 20
)

type formBinding struct{}

func (formBinding) Name() string {
	return "form"
}

func (formBinding) Bind2(req *http.Request, obj interface{}) error {
	if req.Form == nil {
		if err := req.ParseForm(); err != nil {
			return err
		}
	}

	if err := mapFormByTag(obj, req.Form, "form"); err != nil {
		return err
	}

	return validate(obj)
}

func (formBinding) Bind(req *http.Request, obj interface{}) error {
	if req.Form == nil {
		if err := req.ParseForm(); err != nil {
			return err
		}
	}

	if err := MapForm(obj, req.Form, nil, "form"); err != nil {
		return err
	}

	return validate(obj)
}

type formMultipartBinding struct{}

func (formMultipartBinding) Name() string {
	return "multipart/form-data"
}

func (b formMultipartBinding) Bind2(req *http.Request, obj interface{}) error {
	if req.MultipartForm == nil {
		if err := req.ParseMultipartForm(defaultMemory); err != nil {
			return err
		}
	}

	if _, err := mapping(reflect.ValueOf(obj), emptyField, (*multipartForm)(req.MultipartForm), "form"); err != nil {
		return err
	}

	return validate(obj)
}

func (b formMultipartBinding) Bind(req *http.Request, obj interface{}) error {
	if req.MultipartForm == nil {
		if err := req.ParseMultipartForm(defaultMemory); err != nil {
			return err
		}
	}

	if err := MapForm(obj, req.MultipartForm.Value, req.MultipartForm.File, "form"); err != nil {
		return err
	}

	return validate(obj)
}

type formSource map[string][]string

var _ setter = formSource(nil)

// TrySet tries to set a value by request's form source (like map[string][]string)
func (form formSource) TrySet(value reflect.Value, field reflect.StructField, tagValue string, opt setOptions) (isSetted bool, err error) {
	return setByForm(value, field, form, tagValue, opt)
}

type multipartForm multipart.Form

var _ setter = (*multipartForm)(nil)

// TrySet tries to set a value by the multipart request with the binding a form file
func (m *multipartForm) TrySet(value reflect.Value, field reflect.StructField, key string, opt setOptions) (isSetted bool, err error) {
	if files := m.File[key]; len(files) != 0 {
		return setByMultipartFormFile(value, field, files)
	}

	return setByForm(value, field, m.Value, key, opt)
}
