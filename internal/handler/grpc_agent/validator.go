package grpc_agent

import (
	"github.com/a-light-win/pg-helper/pkg/validate"
	"github.com/go-playground/validator/v10"
)

func NewValidator() *validator.Validate {
	Validator := validator.New(validator.WithRequiredStructEnabled())
	validate.RegisterCustomValidations(Validator)
	return Validator
}
