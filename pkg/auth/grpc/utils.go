package grpc

import (
	"context"

	"github.com/a-light-win/pg-helper/pkg/auth"
)

func LoadAuthInfo(ctx context.Context) (*auth.AuthInfo, bool) {
	a, ok := ctx.Value(auth.CtxKeyAuthInfo).(*auth.AuthInfo)
	return a, ok
}
