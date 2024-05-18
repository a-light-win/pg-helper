package gin

import (
	"net/http"

	"github.com/a-light-win/pg-helper/pkg/auth"
	"github.com/gin-gonic/gin"
)

type GinAuth struct {
	*auth.Auth
}

func NewGinAuth(config *auth.AuthConfig) *GinAuth {
	a := auth.NewAuth(config).
		WithAuth(auth.NewJwtAuth(&config.Jwt, jwtTokenFunc)).
		WithAuth(auth.NewMtlsAuth(&config.Mtls, mtlsClientCertFunc))

	return &GinAuth{Auth: a}
}

func (a *GinAuth) AuthMiddleware(c *gin.Context) {
	authInfo, err := a.Parse(c.Request.Context())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.Set(string(auth.CtxKeyAuthInfo), authInfo)

	c.Next()
}

func FetchAuthInfo(c *gin.Context) (*auth.AuthInfo, bool) {
	authInfo, ok := c.Get(string(auth.CtxKeyAuthInfo))
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "auth info not found"})
		return nil, false
	}
	return authInfo.(*auth.AuthInfo), true
}
