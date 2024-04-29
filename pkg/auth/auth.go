package auth

import (
	"context"
	"errors"
)

type AuthApi interface {
	Parse(ctx context.Context) (*AuthInfo, error)
	Enabled() bool
}

type Auth struct {
	Authes []AuthApi
}

func NewAuth(config *AuthConfig) *Auth {
	return &Auth{}
}

func (a *Auth) WithAuth(auth AuthApi) *Auth {
	a.Authes = append(a.Authes, auth)
	return a
}

func (a *Auth) Parse(ctx context.Context) (*AuthInfo, error) {
	var finalAuthInfo *AuthInfo
	for _, auth := range a.Authes {
		if !auth.Enabled() {
			continue
		}
		if authInfo, err := auth.Parse(ctx); err != nil {
			return nil, err
		} else {
			if finalAuthInfo == nil {
				finalAuthInfo = authInfo
				continue
			}
			if err := finalAuthInfo.Merge(authInfo); err != nil {
				return nil, err
			}
		}
	}

	if finalAuthInfo != nil {
		return finalAuthInfo, nil
	}
	return nil, errors.New("no auth method succeeded")
}
