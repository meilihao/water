package binding

import (
	"net/http"
)

type queryBinding struct{}

func (queryBinding) Name() string {
	return "query"
}

func (queryBinding) Bind2(req *http.Request, obj interface{}) error {
	if err := mapFormByTag(obj, req.URL.Query(), "form"); err != nil {
		return err
	}
	return validate(obj)
}

func (queryBinding) Bind(req *http.Request, obj interface{}) error {
	if err := MapForm(obj, req.URL.Query(), nil, "form"); err != nil {
		return err
	}
	return validate(obj)
}
