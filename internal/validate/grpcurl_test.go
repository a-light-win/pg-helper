package validate

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

type TestGrpcUrl struct {
	Url string `validate:"grpcurl"`
}

func TestValidateGrpcUrl(t *testing.T) {
	validate := validator.New()
	_ = validate.RegisterValidation("grpcurl", validateGrpcUrl)

	tests := []struct {
		name  string
		url   string
		valid bool
	}{
		{"Valid DNS", "dns:///example.com:50051", true},
		{"Valid DNS With Server", "dns://1.1.1.1/example.com:50051", true},
		{"Valid IPv4", "ipv4://192.0.2.1:50051", true},
		{"Valid IPv6", "ipv6://[fd03::3]:50051", true},
		{"Valid Paththrough", "passthrough:///example.com:50051", true},
		{"Invalid Scheme", "http://example.com:50051", false},
		{"Missing Port", "dns:///example.com", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validate.Struct(&TestGrpcUrl{Url: test.url})
			if test.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
