package server

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

// Service implements both GrpcService and HttpService
type Service interface {
	GrpcService
	HttpService
}

// Services a list of Services with a name
type Services map[string]Service

// GrpcService is an interface that defines the methods that a gRPC service should implement
type GrpcService interface {
	RegisterGrpcService(gs grpc.ServiceRegistrar)
	PublicGrpcMethods() []string
}

// GrpcServices a list of GrpcService with a name
type GrpcServices map[string]GrpcService

// HttpService is an interface that defines the methods that an HTTP service should implement
type HttpService interface {
	RegisterHttpService(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error
}

// HttpServices a list of HttpService with a name
type HttpServices map[string]HttpService
