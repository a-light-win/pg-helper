package validate

import (
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
)

func RegisterCustomValidations(validatorEngine *validator.Validate) {
	if err := validatorEngine.RegisterValidation("id", validateID); err != nil {
		log.Fatal().Err(err).Msg("Failed to register id validator")
	}
	if err := validatorEngine.RegisterValidation("pg_ver", validatePgVer); err != nil {
		log.Fatal().Err(err).Msg("Failed to register pg_ver validator")
	}
	if err := validatorEngine.RegisterValidation("required_one_if", validateRequiredOneIf); err != nil {
		log.Fatal().Err(err).Msg("Failed to register required_one_if validator")
	}
	if err := validatorEngine.RegisterValidation("grpcurl", validateGrpcUrl); err != nil {
		log.Fatal().Err(err).Msg("Failed to register grpcurl validator")
	}
}

func New() *validator.Validate {
	validatorEngine := validator.New(validator.WithRequiredStructEnabled())
	RegisterCustomValidations(validatorEngine)
	return validatorEngine
}
