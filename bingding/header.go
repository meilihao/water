package binding

import (
	"net/http"
	"net/textproto"
	"reflect"
)

type headerBinding struct{}

func (headerBinding) Name() string {
	return "header"
}

func (headerBinding) Bind(req *http.Request, obj interface{}) error {
	_, err := mapping(reflect.ValueOf(obj), emptyField, headerSource(req.Header), "header")
	if err != nil {
		return err
	}
	return validate(obj)
}

func (headerBinding) Bind2(req *http.Request, obj interface{}) error {
	if err := MapForm(obj, req.Header, nil, "header"); err != nil {
		return err
	}

	return validate(obj)
}

type headerSource map[string][]string

var _ setter = headerSource(nil)

func (hs headerSource) TrySet(value reflect.Value, field reflect.StructField, tagValue string, opt setOptions) (isSetted bool, err error) {
	return setByForm(value, field, hs, textproto.CanonicalMIMEHeaderKey(tagValue), opt)
}
