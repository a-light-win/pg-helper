package grpc

import (
	"context"

	"github.com/a-light-win/pg-helper/pkg/auth"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GrpcAuth struct {
	*auth.Auth
}

func NewGrpcAuth(config *auth.AuthConfig) *GrpcAuth {
	a := auth.NewAuth(config).
		WithAuth(auth.NewJwtAuth(&config.Jwt, JwtTokenFunc)).
		WithAuth(auth.NewMtlsAuth(&config.Mtls, MtlsClientCertFunc))

	return &GrpcAuth{Auth: a}
}

func (a *GrpcAuth) Interceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	authInfo, err := a.Parse(ctx)
	if err != nil {
		return nil, wrappedError(err)
	}

	ctx = context.WithValue(ctx, auth.CtxKeyAuthInfo, authInfo)
	return handler(ctx, req)
}

func (a *GrpcAuth) StreamInterceptor(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	ctx := stream.Context()
	authInfo, err := a.Parse(ctx)
	if err != nil {
		return wrappedError(err)
	}
	ctx = context.WithValue(ctx, auth.CtxKeyAuthInfo, authInfo)
	wrapped := grpc_middleware.WrapServerStream(stream)
	wrapped.WrappedContext = ctx
	return handler(srv, wrapped)
}

func wrappedError(err error) error {
	if _, ok := status.FromError(err); !ok {
		return status.Error(codes.Unauthenticated, err.Error())
	}
	return err
}
