// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.

package runtime

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	valid "github.com/asaskevich/govalidator"
	"github.com/pace/bricks/pkg/isotime"
)

func init() {
	valid.CustomTypeTagMap.Set("iso8601", valid.CustomTypeValidator(func(i interface{}, o interface{}) bool {
		switch v := i.(type) {
		case time.Time:
			return true
		case string:
			_, err := isotime.ParseISO8601(v)
			return err == nil
		}

		return false
	}))
}

// ValidateParameters checks the given struct and returns true if the struct
// is valid according to the specification (declared with go-validator struct tags)
// In case of an error, an jsonapi error message will be directly send to the client
func ValidateParameters(w http.ResponseWriter, r *http.Request, data interface{}) bool {
	return ValidateStruct(w, r, data, "parameter")
}

// ValidateRequest checks the given struct and returns true if the struct
// is valid according to the specification (declared with go-validator struct tags)
// In case of an error, an jsonapi error message will be directly send to the client
func ValidateRequest(w http.ResponseWriter, r *http.Request, data interface{}) bool {
	return ValidateStruct(w, r, data, "pointer")
}

// ValidateStruct checks the given struct and returns true if the struct
// is valid according to the specification (declared with go-validator struct tags)
// In case of an error, an jsonapi error message will be directly send to the client
// The passed source is the source for validation errors (e.g. pointer for data or parameter)
func ValidateStruct(w http.ResponseWriter, r *http.Request, data interface{}, source string) bool {
	ok, err := valid.ValidateStruct(data)

	if !ok {
		switch errs := err.(type) {
		case valid.Errors:
			var e Errors
			generateValidationErrors(errs, &e, source)
			WriteError(w, http.StatusUnprocessableEntity, e)
		case error:
			panic(err) // programming error, e.g. not used with struct
		default:
			panic(fmt.Errorf("unhandled error case: %s", err))
		}

		return false
	}

	return true
}

// convert govalidator errors into jsonapi errors
func generateValidationErrors(validErrors valid.Errors, jsonapiErrors *Errors, source string) {
	for _, err := range validErrors {
		switch e := err.(type) {
		case valid.Errors:
			generateValidationErrors(e, jsonapiErrors, source)
		case valid.Error:
			*jsonapiErrors = append(*jsonapiErrors, generateValidationError(e, source))
		default:
			panic(fmt.Errorf("unhandled error case: %s", e))
		}
	}
}

// BUG(vil): the govalidation error has no reference to the
// original StructField. That makes it impossible to generate
// correct pointers.
// Since the actual data structure and the incoming JSON are very
// different, fork and add struct field tags. Add custom tag
// and use a custom tag to produce correct source pointer/parameter.
// https://github.com/pace/bricks/issues/10

// generateValidationError generates a new jsonapi error based
// on the given govalidator error
func generateValidationError(e valid.Error, source string) *Error {
	path := ""
	for _, p := range append(e.Path, e.Name) {
		path += "/" + strings.ToLower(p)
	}

	// params are prefixed with param remove this until above
	// described bug is fixed with this simple string replace
	if source == "parameter" {
		path = strings.Replace(path, "/param", "", 1)
	}

	return &Error{
		Title:  fmt.Sprintf("%s is invalid", e.Name),
		Detail: e.Err.Error(),
		Source: &map[string]interface{}{
			source: path,
		},
	}
}
