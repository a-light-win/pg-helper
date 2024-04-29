package auth

import (
	"context"
	"crypto/x509"
	"strings"
)

type ClientCertFunc func(ctx context.Context) (*x509.Certificate, error)

type MtlsAuth struct {
	Config         *MtlsAuthConfig
	ClientCertFunc ClientCertFunc
}

func NewMtlsAuth(config *MtlsAuthConfig, clientCertFunc ClientCertFunc) *MtlsAuth {
	return &MtlsAuth{
		Config:         config,
		ClientCertFunc: clientCertFunc,
	}
}

func (mtlsAuth *MtlsAuth) Parse(ctx context.Context) (*AuthInfo, error) {
	clientCert, err := mtlsAuth.ClientCertFunc(ctx)
	if err != nil {
		return nil, err
	}

	authInfo := &AuthInfo{
		ScopeEnabled:    mtlsAuth.Config.ScopeEnabled,
		ResourceEnabled: mtlsAuth.Config.ResourceEnabled,
	}

	for _, san := range clientCert.DNSNames {
		if san, ok := mtlsAuth.stripBaseDomain(san); ok {
			authInfo.FromDNSName(san)
		}
	}

	if err := mtlsAuth.Validate(authInfo); err != nil {
		return nil, err
	}

	return nil, nil
}

func (mtlsAuth *MtlsAuth) Enabled() bool {
	return mtlsAuth.Config.Enabled
}

func (mtlsAuth *MtlsAuth) Validate(authInfo *AuthInfo) error {
	if err := authInfo.Validate(); err != nil {
		return err
	}
	return nil
}

func (mtlsAuth *MtlsAuth) stripBaseDomain(san string) (string, bool) {
	if mtlsAuth.Config.BaseDomain == "" {
		return san, true
	}

	baseDomain := "." + mtlsAuth.Config.BaseDomain

	if !strings.HasSuffix(san, baseDomain) {
		return "", false
	}

	return strings.TrimSuffix(san, baseDomain), true
}
