package binding

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

type sliceValidateError []error

func (err sliceValidateError) Error() string {
	var errMsgs []string
	for i, e := range err {
		if e == nil {
			continue
		}
		errMsgs = append(errMsgs, fmt.Sprintf("[%d]: %s", i, e.Error()))
	}
	return strings.Join(errMsgs, "\n")
}

// StructValidater is an interface which needs to be implemented
type StructValidater interface {
	ValidateStruct(interface{}) error

	// Engine returns the underlying validator engine which powers the
	// StructValidator implementation.
	Engine() interface{}

	Name() string
}

// Validator is the default validator which implements the StructValidator
// interface. It uses https://github.com/go-playground/validator/tree/v10.5.0
var Validator StructValidater = NewvVlidatorV10()

var _ StructValidater = &validatorV10{}

type validatorV10 struct {
	validate *validator.Validate
}

func NewvVlidatorV10() *validatorV10 {
	v := &validatorV10{}

	v.validate = validator.New()
	v.validate.SetTagName("binding")

	return v
}

func (v *validatorV10) Name() string {
	return `validator.v10`
}

func (v *validatorV10) Engine() interface{} {
	return v.validate
}

// ValidateStruct receives any kind of type, but only performed struct or pointer to struct type.
func (v *validatorV10) ValidateStruct(obj interface{}) error {
	if obj == nil {
		return fmt.Errorf("nil obj")
	}

	value := reflect.ValueOf(obj)
	switch value.Kind() {
	case reflect.Ptr:
		return v.ValidateStruct(value.Elem().Interface())
	case reflect.Struct:
		return v.validate.Struct(obj)
	case reflect.Slice, reflect.Array:
		count := value.Len()
		validateRet := make(sliceValidateError, 0)

		for i := 0; i < count; i++ {
			if err := v.ValidateStruct(value.Index(i).Interface()); err != nil {
				validateRet = append(validateRet, err)
			}
		}
		if len(validateRet) == 0 {
			return nil
		}

		return validateRet
	default:
		return nil
	}
}
