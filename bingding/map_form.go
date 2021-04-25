package binding

import (
	"fmt"
	"mime/multipart"
	"reflect"
)

// MapForm map form or multipart.FileHeader to target
func MapForm(obj interface{}, form map[string][]string,
	formfile map[string][]*multipart.FileHeader, tag string) error {
	if obj == nil {
		return fmt.Errorf("nil obj")
	}

	formStruct := reflect.ValueOf(obj)
	var bakPtr interface{}

	if formStruct.Kind() != reflect.Ptr {
		return fmt.Errorf("need ptr obj")
	}

	bakPtr = formStruct.Interface()
	formStruct = formStruct.Elem()

	if formStruct.Kind() == reflect.Map &&
		formStruct.Type().Key().Kind() == reflect.String { // formStruct is map[string]xxx
		return setFormMap2(bakPtr, formStruct.Type(), form)
	}

	return nil
	//return _mapForm(formStruct.Elem(), form, formfile, tag)
}

func _mapForm(formStruct reflect.Value, form map[string][]string,
	formfile map[string][]*multipart.FileHeader, tag string) error {
	if formStruct.Kind() == reflect.Ptr {
		formStruct = formStruct.Elem()
	}
	typ := formStruct.Type()

	_ = typ

	return nil
}
