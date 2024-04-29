package gin

import (
	"context"
	"crypto/x509"
	"errors"

	"github.com/gin-gonic/gin"
)

func mtlsClientCertFunc(ctx context.Context) (*x509.Certificate, error) {
	c, ok := ctx.(*gin.Context)
	if !ok {
		return nil, errors.New("context is not a *gin.Context")
	}

	if c.Request == nil {
		return nil, errors.New("request is nil")
	}

	if c.Request.TLS == nil || len(c.Request.TLS.PeerCertificates) == 0 {
		return nil, errors.New("no client certificate provided")
	}

	return c.Request.TLS.PeerCertificates[0], nil
}
