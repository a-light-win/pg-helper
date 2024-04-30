package validate

import (
	"testing"

	"github.com/go-playground/validator/v10"
)

type TestStruct struct {
	Field1 []int `validate:"samelen=Field2"`
	Field2 []int
}

func TestValidateSameLen(t *testing.T) {
	v := validator.New()
	v.RegisterValidation("samelen", validateSameLen)

	testCases := []struct {
		name   string
		fields TestStruct
		want   bool
	}{
		{
			name: "same length",
			fields: TestStruct{
				Field1: []int{1, 2, 3},
				Field2: []int{4, 5, 6},
			},
			want: true,
		},
		{
			name: "different length",
			fields: TestStruct{
				Field1: []int{1, 2, 3},
				Field2: []int{4, 5},
			},
			want: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := v.Struct(tc.fields)
			if tc.want && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
			if !tc.want && err == nil {
				t.Errorf("expected error, got nil")
			}
		})
	}
}
