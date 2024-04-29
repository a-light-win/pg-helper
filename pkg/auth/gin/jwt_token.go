package gin

import (
	"context"
	"errors"

	"github.com/gin-gonic/gin"
)

func jwtTokenFunc(ctx context.Context) (string, error) {
	if c, ok := ctx.(*gin.Context); ok {
		return c.GetHeader("Authorization"), nil
	}
	return "", errors.New("no Authorization header provided")
}
