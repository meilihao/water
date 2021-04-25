package binding

import (
	"fmt"
	"net/http"

	jsoniter "github.com/json-iterator/go"
)

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary
)

type jsonBinding struct{}

func (jsonBinding) Name() string {
	return "json"
}

func (b jsonBinding) Bind(req *http.Request, obj interface{}) error {
	if req == nil || req.Body == nil {
		return fmt.Errorf("invalid request for bind %s", b.Name())
	}
	defer req.Body.Close()

	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(obj); err != nil {
		return err
	}

	return validate(obj)
}
