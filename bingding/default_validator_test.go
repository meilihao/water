// Copyright 2020 Gin Core Team. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package binding

import (
	"errors"
	"testing"
)

func TestSliceValidateError(t *testing.T) {
	tests := []struct {
		name string
		err  sliceValidateError
		want string
	}{
		{"has nil elements", sliceValidateError{errors.New("test error"), nil}, "[0]: test error"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("sliceValidateError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultValidator(t *testing.T) {
	type exampleStruct struct {
		A string `binding:"max=8"`
		B int    `binding:"gt=0"`
	}
	tests := []struct {
		name    string
		v       *validatorV10
		obj     interface{}
		wantErr bool
	}{
		{"validate nil obj", NewvVlidatorV10(), nil, true},
		{"validate int obj", NewvVlidatorV10(), 3, false},
		{"validate struct failed-1", NewvVlidatorV10(), exampleStruct{A: "123456789", B: 1}, true},
		{"validate struct failed-2", NewvVlidatorV10(), exampleStruct{A: "12345678", B: 0}, true},
		{"validate struct passed", NewvVlidatorV10(), exampleStruct{A: "12345678", B: 1}, false},
		{"validate *struct failed-1", NewvVlidatorV10(), &exampleStruct{A: "123456789", B: 1}, true},
		{"validate *struct failed-2", NewvVlidatorV10(), &exampleStruct{A: "12345678", B: 0}, true},
		{"validate *struct passed", NewvVlidatorV10(), &exampleStruct{A: "12345678", B: 1}, false},
		{"validate []struct failed-1", NewvVlidatorV10(), []exampleStruct{{A: "123456789", B: 1}}, true},
		{"validate []struct failed-2", NewvVlidatorV10(), []exampleStruct{{A: "12345678", B: 0}}, true},
		{"validate []struct passed", NewvVlidatorV10(), []exampleStruct{{A: "12345678", B: 1}}, false},
		{"validate []*struct failed-1", NewvVlidatorV10(), []*exampleStruct{{A: "123456789", B: 1}}, true},
		{"validate []*struct failed-2", NewvVlidatorV10(), []*exampleStruct{{A: "12345678", B: 0}}, true},
		{"validate []*struct passed", NewvVlidatorV10(), []*exampleStruct{{A: "12345678", B: 1}}, false},
		{"validate *[]struct failed-1", NewvVlidatorV10(), &[]exampleStruct{{A: "123456789", B: 1}}, true},
		{"validate *[]struct failed-2", NewvVlidatorV10(), &[]exampleStruct{{A: "12345678", B: 0}}, true},
		{"validate *[]struct passed", NewvVlidatorV10(), &[]exampleStruct{{A: "12345678", B: 1}}, false},
		{"validate *[]*struct failed-1", NewvVlidatorV10(), &[]*exampleStruct{{A: "123456789", B: 1}}, true},
		{"validate *[]*struct failed-2", NewvVlidatorV10(), &[]*exampleStruct{{A: "12345678", B: 0}}, true},
		{"validate *[]*struct passed", NewvVlidatorV10(), &[]*exampleStruct{{A: "12345678", B: 1}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.v.ValidateStruct(tt.obj); (err != nil) != tt.wantErr {
				t.Errorf("defaultValidator.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
