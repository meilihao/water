package binding

import (
	"encoding/xml"
	"fmt"
	"net/http"
)

type xmlBinding struct{}

func (xmlBinding) Name() string {
	return "xml"
}

func (b xmlBinding) Bind(req *http.Request, obj interface{}) error {
	if req == nil || req.Body == nil {
		return fmt.Errorf("invalid request for bind %s", b.Name())
	}
	defer req.Body.Close()

	decoder := xml.NewDecoder(req.Body)
	if err := decoder.Decode(obj); err != nil {
		return err
	}

	return validate(obj)
}
