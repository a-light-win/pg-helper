package validate

import (
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
)

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
