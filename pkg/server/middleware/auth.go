package middleware

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type PeerNameKey string

const PeerName PeerNameKey = "peerName"

type MTLSAuthInterceptor struct {
	// publicMethods is a set of gRPC methods that do not require authentication
	publicMethods map[string]struct{}
}

// NewMTLSAuthInterceptor creates a new MTLSAuthInterceptor
// To make some methods public, you must set ClientAuth as tls.VerifyClientCertIfGiven when credentials.NewTLS()
func NewMTLSAuthInterceptor() *MTLSAuthInterceptor {
	return &MTLSAuthInterceptor{
		publicMethods: make(map[string]struct{}),
	}
}

// ToPublic return a function that adds a gRPC method to the public methods set
func (i *MTLSAuthInterceptor) ToPublic() func(string) {
	return func(method string) {
		i.publicMethods[method] = struct{}{}
	}
}

func (i *MTLSAuthInterceptor) IsPublic(method string) bool {
	_, ok := i.publicMethods[method]
	return ok
}

// AuthInterceptor is a gRPC middleware for mTLS authentication
func (i *MTLSAuthInterceptor) AuthInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Check if the method requires authentication
		// localhost is always public for grpc gateway
		if i.IsPublic(info.FullMethod) {
			return handler(ctx, req)
		} else if isLocalhost() {
			ctx = context.WithValue(ctx, PeerName, "localhost")
			return handler(ctx, req)
		}

		// Perform mTLS authentication
		if err := authenticateMTLS(ctx); err != nil {
			return nil, err
		}

		// Proceed with the request
		ctx = context.WithValue(ctx, PeerName, getPeerName(ctx))
		return handler(ctx, req)
	}
}

func isLocalhost() bool {
	return true
}

// authenticateMTLS performs mTLS authentication
func authenticateMTLS(ctx context.Context) error {
	// 从 gRPC 的 context 获取 peer 信息
	p, ok := peer.FromContext(ctx)
	if !ok {
		return status.Errorf(codes.Unauthenticated, "unable to retrieve peer information")
	}

	tlsInfo, ok := p.AuthInfo.(credentials.TLSInfo)
	if !ok {
		return status.Errorf(codes.Unauthenticated, "you must use mTLS to access this method")
	}

	if len(tlsInfo.State.VerifiedChains) == 0 && len(tlsInfo.State.VerifiedChains[0]) == 0 {
		return status.Errorf(codes.Unauthenticated, "client certificate verification failed")
	}

	return nil
}

func getPeerName(ctx context.Context) string {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return ""
	}

	tlsInfo, ok := p.AuthInfo.(credentials.TLSInfo)
	if !ok {
		return ""
	}

	if len(tlsInfo.State.VerifiedChains) == 0 {
		return ""
	}

	return tlsInfo.State.VerifiedChains[0][0].Subject.CommonName
}
