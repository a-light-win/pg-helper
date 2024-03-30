package handler

import (
	"regexp"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
)

var id_re = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]+$`)

func validateID(fl validator.FieldLevel) bool {
	return id_re.MatchString(fl.Field().String())
}

func init() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		if err := v.RegisterValidation("id", validateID); err != nil {
			log.Error().Err(err).Msg("Failed to register id validator")
		}
	}
}
