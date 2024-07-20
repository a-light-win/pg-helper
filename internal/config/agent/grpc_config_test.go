package agent

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGrpcClientConfig_AuthToken(t *testing.T) {
	// Create a temporary file
	tempFile, err := os.CreateTemp("", "pg_helper_test-grpc_config")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempFile.Name())

	// Write a test token to the file
	testToken := "test-token \n"
	if _, err := io.WriteString(tempFile, testToken); err != nil {
		t.Fatal(err)
	}

	// Create a GrpcClientConfig with the temporary file as the AuthTokenFile
	config := &GrpcClientConfig{
		AuthTokenFile: tempFile.Name(),
	}

	// Call AuthToken and check the returned token
	token, err := config.AuthToken()
	assert.NoError(t, err)
	assert.Equal(t, "test-token", token)
}
