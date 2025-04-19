package server

import (
	"xdp-banner/agent/internal/service"
	"xdp-banner/agent/internal/statusfsm"
	"xdp-banner/pkg/server"

	"google.golang.org/grpc/credentials"
)

type GrpcServer struct {
	*server.GrpcServer
}

func NewGrpcServer(services server.GrpcServices, creds credentials.TransportCredentials) GrpcServer {
	grpcServer := server.NewGrpcServer(services, creds)

	return GrpcServer{
		GrpcServer: &grpcServer,
	}
}

func NewGrpcServices(fsm *statusfsm.StatusFSM) server.GrpcServices {
	return server.GrpcServices{
		"control": service.NewControlService(fsm),
	}
}
