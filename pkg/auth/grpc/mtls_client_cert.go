package grpc

import (
	"context"
	"crypto/x509"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

func MtlsClientCertFunc(ctx context.Context) (*x509.Certificate, error) {
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
	return tlsInfo.State.PeerCertificates[0], nil
}
