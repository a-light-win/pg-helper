package gin

import (
	"context"

	"github.com/a-light-win/pg-helper/pkg/auth"
	"github.com/gin-gonic/gin"
)

func LoadAuthInfo(ctx context.Context) (*auth.AuthInfo, bool) {
	if c, ok := ctx.(*gin.Context); ok {
		rawValue, ok := c.Get(string(auth.CtxKeyAuthInfo))
		if ok {
			a, ok := rawValue.(*auth.AuthInfo)
			return a, ok
		}
	}
	return nil, false
}
