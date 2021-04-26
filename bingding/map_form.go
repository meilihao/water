package binding

import (
	"fmt"
	"mime/multipart"
	"net/textproto"
	"reflect"
	"strconv"
	"time"
)

// MapForm map form or multipart.FileHeader to target
// obj support map[string]string, map[string][]string, struct
// struct field指针超过一层没有意义, 不支持
func MapForm(obj interface{}, form map[string][]string,
	formfile map[string][]*multipart.FileHeader, tag string) error {
	if obj == nil {
		return fmt.Errorf("nil obj")
	}

	formStruct := reflect.ValueOf(obj)

	if formStruct.Kind() != reflect.Ptr {
		return fmt.Errorf("need ptr obj")
	}

	formStruct = formStruct.Elem()
	typ := formStruct.Type()

	if formStruct.Kind() == reflect.Map &&
		typ.Key().Kind() == reflect.String { // formStruct is map[string]xxx
		return setFormMap2(obj, typ, form)
	}

	if typ.Kind() != reflect.Struct {
		return fmt.Errorf("form mapping need map[string]string, map[string][]string or struct")
	}

	return _mapForm(formStruct, form, formfile, tag)
}

// func _redirectValue(typ reflect.Type, formStruct reflect.Value) (reflect.Type, reflect.Value) {
// 	for {
// 		if typ.Kind() == reflect.Ptr {
// 			break
// 		}

// 			if formStruct.IsNil() {
// 				formStruct = reflect.New(typ.Elem()).Elem()
// 			} else {
// 				formStruct = formStruct.Elem()
// 			}

// 			typ = formStruct.Type()
// 	}

// 	return
// }

func _mapForm(formStruct reflect.Value, form map[string][]string,
	formfile map[string][]*multipart.FileHeader, tag string) error {
	if formStruct.Kind() == reflect.Ptr {
		formStruct = formStruct.Elem()
	}
	typ := formStruct.Type()

	for i := 0; i < typ.NumField(); i++ {
		typeField := typ.Field(i)
		structField := formStruct.Field(i)

		if typeField.PkgPath != "" && !typeField.Anonymous { // unexported
			continue
		}

		if typeField.Type.Kind() == reflect.Ptr && typeField.Type != multipartFileType {
			if structField.IsNil() && (typeField.Anonymous || typeField.Type.Elem().Kind() == reflect.Struct) {
				structField.Set(reflect.New(typeField.Type.Elem()))

				structField = structField.Elem()
			}
		}

		if structField.Kind() == reflect.Struct || typeField.Anonymous { // typeField.Anonymous is an embedded field
			if err := _mapForm(structField, form, formfile, tag); err != nil {
				return err
			}
		}

		if err := tryToSetValue2(structField, typeField, form, formfile, tag); err != nil {
			return err
		}
	}

	return nil
}

var (
	multipartFileType = reflect.TypeOf((*multipart.FileHeader)(nil))
)

// typeField does't use typeField.Type.Kind(), typeField only form tag in here
func tryToSetValue2(value reflect.Value, typeField reflect.StructField, form map[string][]string,
	formfile map[string][]*multipart.FileHeader, tag string) error {

	bName, bOpt := parseFormName(typeField, tag)
	if bName == "-" {
		return nil
	}

	inputValue, existValue := form[bName]
	if !existValue && bOpt.defaultValue != "" {
		existValue = true
		inputValue = []string{bOpt.defaultValue}
	}
	if existValue {
		var err error
		num := len(inputValue)

		if value.Kind() == reflect.Slice && num > 0 {
			sliceOf := value.Type().Elem().Kind()
			slice := reflect.MakeSlice(value.Type(), num, num)
			for i := 0; i < num; i++ {
				if err = setWithProperType2(typeField, bOpt, sliceOf, inputValue[i], slice.Index(i), bName); err != nil {
					return err
				}
			}
			value.Set(slice)
		} else {
			err = setWithProperType2(typeField, bOpt, value.Kind(), inputValue[0], value, bName)
		}

		return err
	}

	inputFile, existFile := formfile[bName]
	if !existFile {
		return nil
	}
	num := len(inputFile)

	// value is []*multipart.FileHeader
	if value.Kind() == reflect.Slice && num > 0 && value.Type().Elem() == multipartFileType {
		slice := reflect.MakeSlice(value.Type(), num, num)
		for i, n := 0, len(inputFile); i < n; i++ {
			slice.Index(i).Set(reflect.ValueOf(inputFile[i]))
			num--
		}
		value.Set(slice)
	} else if value.Type() == multipartFileType { // value is *multipart.FileHeader
		value.Set(reflect.ValueOf(inputFile[0]))
		num--
	}

	if num != len(inputFile) { //success mapping
		return nil
	}

	return fmt.Errorf("%s can't mapping to multipart.FileHeader", bName)
}

func setValueError(name, value string) error {
	return fmt.Errorf("can't set %s with %s", name, value)
}

func setWithProperType2(typeField reflect.StructField, opt *setOptions, valueKind reflect.Kind, val string, structField reflect.Value, nameInTag string) error {
	if structField.Kind() == reflect.Ptr {
		if structField.IsNil() {
			structField.Set(reflect.New(structField.Type().Elem()))
		}

		structField = structField.Elem()
		valueKind = structField.Kind()
	}

	switch valueKind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if val == "" {
			val = "0"
		}
		intVal, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return setValueError(nameInTag, val)
		} else {
			structField.SetInt(intVal)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if val == "" {
			val = "0"
		}
		uintVal, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return setValueError(nameInTag, val)
		} else {
			structField.SetUint(uintVal)
		}
	case reflect.Bool:
		if val == "" {
			val = "false"
		}
		boolVal, err := strconv.ParseBool(val)
		if err != nil {
			return setValueError(nameInTag, val)
		} else if boolVal {
			structField.SetBool(true)
		}
	case reflect.Float32:
		if val == "" {
			val = "0.0"
		}
		floatVal, err := strconv.ParseFloat(val, 32)
		if err != nil {
			return setValueError(nameInTag, val)
		} else {
			structField.SetFloat(floatVal)
		}
	case reflect.Float64:
		if val == "" {
			val = "0.0"
		}
		floatVal, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return setValueError(nameInTag, val)
		} else {
			structField.SetFloat(floatVal)
		}
	case reflect.String:
		structField.SetString(val)
	case reflect.Struct:
		switch structField.Interface().(type) {
		case time.Time:
			return setTimeField2(val, opt, structField)
		}
		return json.Unmarshal([]byte(val), structField.Addr().Interface())
	case reflect.Map:
		return json.Unmarshal([]byte(val), structField.Addr().Interface())
	default:
		return errUnknownType
	}

	return nil
}

func parseFormName(typeField reflect.StructField, tag string) (string, *setOptions) {
	setOpt := &setOptions{}
	setOpt.timeFormat = typeField.Tag.Get("time_format")
	setOpt.timeUtc = typeField.Tag.Get("time_utc")
	setOpt.timeLocation = typeField.Tag.Get("time_location")

	name, tail := head(typeField.Tag.Get(tag), ",")

	if name == "" { // default value is FieldName
		name = "-"
	}
	if tag == "header" && name != "-" {
		name = textproto.CanonicalMIMEHeaderKey(name)
	}

	var opt string
	for len(tail) > 0 {
		opt, tail = head(tail, ",")

		if k, v := head(opt, "="); k == "default" {
			setOpt.defaultValue = v
		}
	}

	return name, setOpt
}
