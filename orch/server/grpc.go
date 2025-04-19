package server

import (
	"xdp-banner/pkg/log"
	"xdp-banner/pkg/server"
)

type GrpcServer struct {
	*server.GrpcServer
}

func NewGrpcServer(services server.Services) GrpcServer {
	creds, err := NewCredits()
	if err != nil {
		log.Fatal("failed to create credentials", log.ErrorField(err))
	}

	// convert services to GrpcServices
	grpcServices := make(server.GrpcServices, len(services))
	for name, service := range services {
		grpcServices[name] = service
	}

	grpcServer := server.NewGrpcServer(grpcServices, creds)

	return GrpcServer{
		GrpcServer: &grpcServer,
	}
}
