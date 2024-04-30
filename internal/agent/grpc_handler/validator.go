package grpc_handler

import (
	"github.com/a-light-win/pg-helper/pkg/validate"
	"github.com/go-playground/validator/v10"
)

var Validator *validator.Validate

func InitValidator() {
	Validator = validator.New(validator.WithRequiredStructEnabled())
	validate.RegisterCustomValidations(Validator)
}
