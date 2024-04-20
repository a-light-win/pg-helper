package grpc_handler

import "context"

type AuthToken struct {
	token  string
	secure bool
}

func (a *AuthToken) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": "Bearer " + a.token,
	}, nil
}

func (a *AuthToken) RequireTransportSecurity() bool {
	return a.secure
}

func NewAuthToken(token string, secure bool) *AuthToken {
	return &AuthToken{
		token:  token,
		secure: secure,
	}
}
