package auth

import (
	"context"
	"errors"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type JwtTokenFunc func(ctx context.Context) (string, error)

type JwtAuth struct {
	Config    *JwtAuthConfig
	TokenFunc JwtTokenFunc
}

func NewJwtAuth(config *JwtAuthConfig, tokenFunc JwtTokenFunc) *JwtAuth {
	return &JwtAuth{
		Config: config,
	}
}

func (jwtAuth *JwtAuth) Parse(ctx context.Context) (*AuthInfo, error) {
	tokenStr, err := jwtAuth.getToken(ctx)
	if err != nil {
		return nil, err
	}

	token, err := jwt.Parse(tokenStr, jwtAuth.Config.LoadVerifyKey)
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		authInfo := &AuthInfo{
			ScopeEnabled:    true,
			ResourceEnabled: true,
		}

		authInfo.FromClaims(claims)

		if err := jwtAuth.Validate(authInfo); err != nil {
			return nil, err
		}

		return authInfo, nil
	} else {
		return nil, errors.New("invalid token")
	}
}

func (jwtAuth *JwtAuth) Enabled() bool {
	return jwtAuth.Config.Enabled
}

func (jwtAuth *JwtAuth) Validate(authInfo *AuthInfo) error {
	if err := authInfo.Validate(); err != nil {
		return err
	}

	if jwtAuth.Config.Audience != authInfo.Audience {
		return errors.New("invalid audience")
	}
	return nil
}

func (jwtAuth *JwtAuth) getToken(ctx context.Context) (string, error) {
	authHeader, err := jwtAuth.TokenFunc(ctx)
	if err != nil {
		return "", err
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", errors.New("invalid authorization header")
	}

	return strings.TrimPrefix(authHeader, "Bearer "), nil
}
