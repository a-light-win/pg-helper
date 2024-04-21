package grpc_handler

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type (
	CtxKey     string
	ClientType string
)

const (
	CtxKeyAuthInfo CtxKey = "AuthInfo"
)

const (
	AgentClient ClientType = "agent"
	AppClient   ClientType = "app"
)

type AuthInfo struct {
	ClientId   string
	ClientType ClientType
	BaseDomain string
}

func (a AuthInfo) ClientTypeString() string {
	return string(a.ClientType)
}

func TlsAuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Get the credentials from the context
	p, ok := peer.FromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "no transport security info available")
	}

	// Get the TLS info
	tlsInfo, ok := p.AuthInfo.(credentials.TLSInfo)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "unable to get TLS info")
	}

	// Get the client's certificate
	clientCert := tlsInfo.State.PeerCertificates[0]

	// Get the CN
	sans := clientCert.DNSNames

	for _, san := range sans {
		authInfo, ok := parseAuthFromCert(san)
		if ok {
			ctx = context.WithValue(ctx, CtxKeyAuthInfo, authInfo)
			return handler(ctx, req)
		}
	}

	// Call the handler function with the new context
	return nil, status.Errorf(codes.Unauthenticated, "invalid certificate")
}

func BearerAuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "metadata is not provided")
	}

	authHeader, ok := md["authorization"]
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "authorization header is not provided")
	}

	token := strings.TrimPrefix(authHeader[0], "Bearer ")
	authInfo, ok := parseAuthFromToken(token)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token")
	}

	ctx = context.WithValue(ctx, CtxKeyAuthInfo, authInfo)
	return handler(ctx, req)
}

func parseAuthFromCert(san string) (*AuthInfo, bool) {
	// Parse the agentId and pgVersion from the CN
	// The CN should be in the format of "agentId.pgVersion.example.com"

	names := strings.SplitN(san, ".", 3)
	if len(names) != 3 {
		return nil, false
	}

	authInfo := AuthInfo{
		ClientId:   names[0],
		ClientType: ClientType(names[1]),
		BaseDomain: names[2],
	}
	if err := validAuthInfo(authInfo); err != nil {
		log.Warn().Err(err).
			Str("ClientId", authInfo.ClientId).
			Str("ClientType", string(authInfo.ClientType)).
			Str("BaseDomain", authInfo.BaseDomain).
			Msg("")
		return nil, false
	}

	return &authInfo, true
}

func parseAuthFromToken(tokenString string) (*AuthInfo, bool) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		switch token.Method.(type) {
		case *jwt.SigningMethodECDSA:
			return gd_.GrpcConfig.JwtEcDSAVerifyKey()
		case *jwt.SigningMethodEd25519:
			return gd_.GrpcConfig.JwtEdDSAVerifyKey()
		default:
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
	})
	if err != nil {
		return nil, false
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		authInfo := AuthInfo{
			ClientId:   claims["sub"].(string),
			ClientType: ClientType(claims["type"].(string)),
			BaseDomain: claims["baseDomain"].(string),
		}
		if err := validAuthInfo(authInfo); err != nil {
			return nil, false
		}
		return &authInfo, true
	} else {
		return nil, false
	}
}

func validAuthInfo(authInfo AuthInfo) error {
	if gd_.GrpcConfig.Tls.TrustedClientDomain != "" {
		if gd_.GrpcConfig.Tls.TrustedClientDomain != authInfo.BaseDomain {
			return errors.New("invalid base domain")
		}
	}
	switch authInfo.ClientType {
	case AgentClient:
		return nil
	case AppClient:
		return nil
	default:
		return errors.New("invalid client type")
	}
}

func authInfoWithType(ctx context.Context, clientType ClientType) (*AuthInfo, error) {
	authInfo, ok := ctx.Value(CtxKeyAuthInfo).(*AuthInfo)
	if !ok {
		return nil, errors.New("authInfo not found")
	}
	if authInfo.ClientType != clientType {
		return nil, errors.New("invalid client type")
	}
	return authInfo, nil
}
