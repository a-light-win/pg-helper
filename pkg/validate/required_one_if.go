package validate

import (
	"strings"

	"github.com/go-playground/validator/v10"
)

// The required_one_if validator checks if current field value
// is equal to the first param, if so, one of the given fields should not be empty
//
// Example:
//
//	type Example struct {
//		Enabled bool `validate:"required_one_if=true Field1 Filed2"`
//		Field1 string
//		Field2 string
//	}
func validateRequiredOneIf(field validator.FieldLevel) bool {
	// Get the value of the field
	value := field.Field().String()

	// Get the name of the tag
	tagName := field.Param()
	tags := strings.Split(tagName, " ")

	expectValue := tags[0]
	tags = tags[1:]

	if value != expectValue {
		return true
	}

	parent := field.Parent()
	for _, tagName := range tags {
		tagValue := parent.FieldByName(tagName).String()
		if tagValue != "" {
			return true
		}
	}
	return false
}
