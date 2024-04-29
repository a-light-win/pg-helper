package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc/metadata"
)

func JwtTokenFunc(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("metadata not found")
	}

	if len(md["authorization"]) == 0 {
		return "", errors.New("no Authorization header provided")
	}
	return md["authorization"][0], nil
}
