package validate

import (
	"strings"

	"github.com/go-playground/validator/v10"
)

func validateSameLen(field validator.FieldLevel) bool {
	value := field.Field()
	params := strings.Split(field.Param(), " ")

	for _, param := range params {
		otherField := field.Parent().FieldByName(param)
		if value.Len() != otherField.Len() {
			return false
		}
	}

	return true
}
