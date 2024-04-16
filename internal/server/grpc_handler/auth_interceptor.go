package grpc_handler

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
)

type CtxKey string

const (
	CtxKeyAuthInfo CtxKey = "AuthInfo"
)

type AuthInfo struct {
	AgentId   string
	PgVersion int32
}

func TlsAuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Get the credentials from the context
	p, ok := peer.FromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("no transport security set")
	}

	// Get the TLS info
	tlsInfo, ok := p.AuthInfo.(credentials.TLSInfo)
	if !ok {
		return nil, fmt.Errorf("unable to get TLS info")
	}

	// Get the client's certificate
	clientCert := tlsInfo.State.PeerCertificates[0]

	// Get the CN
	sans := clientCert.DNSNames

	for _, san := range sans {
		authInfo, ok := parseAgentId(san)
		if ok {
			ctx = context.WithValue(ctx, CtxKeyAuthInfo, authInfo)
			return handler(ctx, req)
		}
	}

	// Call the handler function with the new context
	return nil, errors.New("invalid certificate")
}

func parseAgentId(san string) (AuthInfo, bool) {
	// Parse the agentId and pgVersion from the CN
	// The CN should be in the format of "agentId.pgVersion.example.com"

	names := strings.Split(san, ".")
	if len(names) < 4 {
		return AuthInfo{}, false
	}

	agentId := names[0]
	pgVersion, err := strconv.Atoi(names[1])
	if err != nil {
		return AuthInfo{}, false
	}

	return AuthInfo{AgentId: agentId, PgVersion: int32(pgVersion)}, true
}
