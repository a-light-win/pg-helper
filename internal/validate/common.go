package validate

import (
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
)

func RegisterCustomValidations(validatorEngine *validator.Validate) {
	// Register the "id" validation function.
	if err := validatorEngine.RegisterValidation("id", validateID); err != nil {
		log.Fatal().Err(err).Msg("Failed to register id validator")
	}
	if err := validatorEngine.RegisterValidation("pg_ver", validatePgVer); err != nil {
		log.Fatal().Err(err).Msg("Failed to register pg_ver validator")
	}
}
