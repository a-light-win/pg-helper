package validate

import "github.com/go-playground/validator/v10"

var MinSupportedPgVersion int64 = 13

func validatePgVer(field validator.FieldLevel) bool {
	value := field.Field()
	if value.IsZero() {
		return true
	}
	return value.Int() >= MinSupportedPgVersion
}
