package server

import (
	"fmt"
	"net"
	"xdp-banner/pkg/log"
	"xdp-banner/pkg/server/middleware"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type ToPublicMethod func(fullMethodName string)

type GrpcServer struct {
	server    *grpc.Server
	authInter *middleware.MTLSAuthInterceptor
}

func NewGrpcServer(service GrpcServices, creds credentials.TransportCredentials) GrpcServer {
	authInter := middleware.NewMTLSAuthInterceptor()

	gs := grpc.NewServer(
		grpc.Creds(creds),
		grpc.UnaryInterceptor(authInter.AuthInterceptor()),
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)

	setupGrpcService(service, gs)
	setupGrpcPublicMethod(service, authInter.ToPublic())

	return GrpcServer{
		server:    gs,
		authInter: authInter,
	}
}

func (gs GrpcServer) Serve(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}
	log.Info("Starting gRPC server on port " + addr)

	return gs.server.Serve(lis)
}

func (gs GrpcServer) Stop() {
	gs.server.Stop()
}

func setupGrpcService(service GrpcServices, gs grpc.ServiceRegistrar) {
	for _, s := range service {
		s.RegisterGrpcService(gs)
	}
}

func setupGrpcPublicMethod(service GrpcServices, toPublicMethod ToPublicMethod) {
	for _, s := range service {
		for _, m := range s.PublicGrpcMethods() {
			toPublicMethod(m)
		}
	}
}
