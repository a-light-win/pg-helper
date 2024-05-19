package validate

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

var validNamePattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_-]+$`)

func isValidName(field validator.FieldLevel) bool {
	if field.Field().String() == "" {
		return true
	}
	return validNamePattern.MatchString(field.Field().String())
}
