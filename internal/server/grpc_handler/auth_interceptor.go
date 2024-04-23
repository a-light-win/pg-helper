package grpc_handler

import (
	"context"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
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
	Subject    string
	ClientType ClientType
	BaseDomain string
}

func (a AuthInfo) ClientTypeString() string {
	return string(a.ClientType)
}

func (a *AuthInfo) FromClaims(claims jwt.MapClaims) {
	if subject, ok := claims["sub"].(string); ok {
		a.Subject = subject
	}
	if clientType, ok := claims["type"].(string); ok {
		a.ClientType = ClientType(clientType)
	}
	if baseDomain, ok := claims["base_domain"].(string); ok {
		a.BaseDomain = baseDomain
	}
}

func (a *AuthInfo) Validate() error {
	if a.Subject == "" {
		return status.Error(codes.Unauthenticated, "invalid subject")
	}

	if gd_.GrpcConfig.Tls.TrustedClientDomain != "" {
		if gd_.GrpcConfig.Tls.TrustedClientDomain != a.BaseDomain {
			return status.Error(codes.Unauthenticated, "invalid base domain")
		}
	}

	switch a.ClientType {
	case AgentClient:
		return nil
	case AppClient:
		return nil
	default:
		return status.Error(codes.Unauthenticated, "invalid client type")
	}
}

func AuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	ctx, err := withAuthInfo(ctx)
	if err != nil {
		return nil, err
	}

	return handler(ctx, req)
}

func AuthStreamInterceptor(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	ctx := stream.Context()
	ctx, err := withAuthInfo(ctx)
	if err != nil {
		return err
	}

	wrapped := grpc_middleware.WrapServerStream(stream)
	wrapped.WrappedContext = ctx
	return handler(srv, wrapped)
}

func withAuthInfo(ctx context.Context) (context.Context, error) {
	var firstErr error
	if gd_.GrpcConfig.BearerAuthEnabled {
		if ctx, err := withAuthInfoFromHeader(ctx); err == nil {
			return ctx, nil
		} else {
			firstErr = err
		}
	}

	if gd_.GrpcConfig.Tls.MTLSEnabled {
		if ctx, err := withAuthInfoFromTls(ctx); err == nil {
			return ctx, nil
		} else if firstErr == nil {
			firstErr = err
		}
	}

	if firstErr != nil {
		return nil, firstErr
	}
	return nil, status.Error(codes.Unauthenticated, "no auth method found")
}

func withAuthInfoFromTls(ctx context.Context) (context.Context, error) {
	// Get the credentials from the context
	p, ok := peer.FromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "no transport security info available")
	}

	// Get the TLS info
	tlsInfo, ok := p.AuthInfo.(credentials.TLSInfo)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unable to get TLS info")
	}

	// Get the client's certificate
	if len(tlsInfo.State.PeerCertificates) == 0 {
		return nil, status.Error(codes.Unauthenticated, "no client certificate provided")
	}
	clientCert := tlsInfo.State.PeerCertificates[0]

	// Get the CN
	sans := clientCert.DNSNames
	for _, san := range sans {
		authInfo, ok := parseAuthFromCert(san)
		if ok {
			ctx = context.WithValue(ctx, CtxKeyAuthInfo, authInfo)
			return ctx, nil
		}
	}

	// Call the handler function with the new context
	return nil, status.Error(codes.Unauthenticated, "invalid certificate")
}

func withAuthInfoFromHeader(ctx context.Context) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "metadata is not provided")
	}

	authHeader, ok := md["authorization"]
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "authorization header is not provided")
	}

	token := strings.TrimPrefix(authHeader[0], "Bearer ")
	authInfo, err := parseAuthFromToken(token)
	if err != nil {
		return nil, err
	}

	ctx = context.WithValue(ctx, CtxKeyAuthInfo, authInfo)
	return ctx, nil
}

func parseAuthFromCert(san string) (*AuthInfo, bool) {
	// Parse the agentId and pgVersion from the CN
	// The CN should be in the format of "agentId.pgVersion.example.com"

	names := strings.SplitN(san, ".", 3)
	if len(names) != 3 {
		return nil, false
	}

	authInfo := &AuthInfo{
		Subject:    names[0],
		ClientType: ClientType(names[1]),
		BaseDomain: names[2],
	}
	if err := authInfo.Validate(); err != nil {
		log.Warn().Err(err).
			Str("ClientId", authInfo.Subject).
			Str("ClientType", string(authInfo.ClientType)).
			Str("BaseDomain", authInfo.BaseDomain).
			Msg("")
		return nil, false
	}

	return authInfo, true
}

func parseAuthFromToken(tokenString string) (*AuthInfo, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		switch token.Method.(type) {
		case *jwt.SigningMethodECDSA:
			return gd_.GrpcConfig.JwtEcDSAVerifyKey()
		case *jwt.SigningMethodEd25519:
			return gd_.GrpcConfig.JwtEdDSAVerifyKey()
		default:
			return nil, status.Errorf(codes.Unauthenticated, "unexpected signing method: %v", token.Header["alg"])
		}
	})
	if err != nil {
		if _, ok := status.FromError(err); !ok {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		authInfo := &AuthInfo{}
		authInfo.FromClaims(claims)
		if err := authInfo.Validate(); err != nil {
			return nil, err
		}
		return authInfo, nil
	} else {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}
}

func authInfoWithType(ctx context.Context, clientType ClientType) (*AuthInfo, error) {
	authInfo, ok := ctx.Value(CtxKeyAuthInfo).(*AuthInfo)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "authInfo not found")
	}
	if authInfo.ClientType != clientType {
		return nil, status.Error(codes.Unauthenticated, "invalid client type")
	}
	return authInfo, nil
}
