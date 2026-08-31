package utils

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()

	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

// FormatValidationErrors formats raw validator errors into a client-friendly map
func FormatErrors(err error) map[string]string {
	errs := make(map[string]string)

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			field := e.Field()
			switch e.Tag() {
			case "required":
				errs[field] = fmt.Sprintf("%s is required", field)
			case "email":
				errs[field] = fmt.Sprintf("%s must be a valid email address", field)
			case "min":
				errs[field] = fmt.Sprintf("%s must be at least %s characters", field, e.Param())
			case "max":
				errs[field] = fmt.Sprintf("%s must not exceed %s characters", field, e.Param())
			case "oneof":
				errs[field] = fmt.Sprintf("%s must be one of [%s]", field, e.Param())
			default:
				errs[field] = fmt.Sprintf("%s failed on rule '%s'", field, e.Tag())
			}
		}
	}

	return errs
}

func ValidateStruct(s any) error {
	return validate.Struct(s)
}
