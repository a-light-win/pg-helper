package validate

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

// idRegex matches valid IDs.
var idRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]+$`)

// validateID checks if the field value matches the idRegex.
func validateID(field validator.FieldLevel) bool {
	if field.Field().String() == "" {
		return true
	}
	return idRegex.MatchString(field.Field().String())
}
