package binding

import (
	"net/http"
)

type queryBinding struct{}

func (queryBinding) Name() string {
	return "query"
}

func (queryBinding) Bind(req *http.Request, obj interface{}) error {
	if err := mapFormByTag(obj, req.URL.Query(), "form"); err != nil {
		return err
	}
	return validate(obj)
}
