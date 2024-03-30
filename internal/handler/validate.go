package handler

import (
	"regexp"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
)

// idRegex matches valid IDs.
var idRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]+$`)

// validateID checks if the field value matches the idRegex.
func validateID(field validator.FieldLevel) bool {
	return idRegex.MatchString(field.Field().String())
}

func RegisterCustomValidations() {
	// Get the validator engine.
	validatorEngine, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		log.Fatal().Msg("Failed to get validator engine")
	}

	// Register the "id" validation function.
	if err := validatorEngine.RegisterValidation("id", validateID); err != nil {
		log.Fatal().Err(err).Msg("Failed to register id validator")
	}
}
