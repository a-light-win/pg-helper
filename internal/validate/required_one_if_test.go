package validate

import (
	"testing"

	"github.com/go-playground/validator/v10"
)

type TestRequiredOneIf struct {
	Field1 string `validate:"required_oneof_if=value1 Field3 Field4"`
	Field2 string
	Field3 string
	Field4 string
}

func TestValidateRequiredOneIf(t *testing.T) {
	validate := validator.New(validator.WithRequiredStructEnabled())
	validate.RegisterValidation("required_oneof_if", validateRequiredOneIf)

	tests := []struct {
		name  string
		value TestRequiredOneIf
		want  bool
	}{
		{
			name: "Field1 not equal to value1",
			value: TestRequiredOneIf{
				Field1: "value2",
				Field2: "value1",
				Field3: "",
				Field4: "",
			},
			want: true,
		},
		{
			name: "Field1 equal to value1, Field3 not empty",
			value: TestRequiredOneIf{
				Field1: "value1",
				Field2: "value1",
				Field3: "value2",
				Field4: "",
			},
			want: true,
		},
		{
			name: "Field1 equal to value1, all other fields empty",
			value: TestRequiredOneIf{
				Field1: "value1",
				Field2: "value1",
				Field3: "",
				Field4: "",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Struct(tt.value)
			isSuccess := err == nil
			wantSuccess := tt.want
			if isSuccess != wantSuccess {
				t.Errorf("validateRequiredOneOfIf() error = %v, wantErr %v", err, tt.want)
			}
		})
	}
}
