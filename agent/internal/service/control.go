package service

import (
	"context"
	"xdp-banner/agent/internal/statusfsm"
	"xdp-banner/api/agent/v1/control"
	"xdp-banner/pkg/log"
	"xdp-banner/pkg/server/common"

	"google.golang.org/grpc"
)

type ControlService struct {
	fsm *statusfsm.StatusFSM
	control.UnimplementedControlServiceServer
}

func NewControlService(fsm *statusfsm.StatusFSM) *ControlService {
	return &ControlService{
		fsm: fsm,
	}
}

func (s *ControlService) RegisterGrpcService(gs grpc.ServiceRegistrar) {
	control.RegisterControlServiceServer(gs, s)
}

func (s *ControlService) PublicGrpcMethods() []string {
	return nil
}

func (s *ControlService) Start(ctx context.Context, req *control.StartRequest) (*control.StartResponse, error) {
	if req == nil {
		return nil, common.InvalidArgumentError("request is nil")
	} else if req.ConfigName == "" {
		return nil, common.InvalidArgumentError("missing config name")
	}

	log.Debug("Received start request", log.StringField("config_name", req.ConfigName))
	s.fsm.Event(statusfsm.Start, req.ConfigName)
	return &control.StartResponse{}, nil
}

func (s *ControlService) Stop(ctx context.Context, req *control.StopRequest) (*control.StopResponse, error) {
	if req == nil {
		return nil, common.InvalidArgumentError("request is nil")
	}

	log.Debug("Received stop request")

	s.fsm.Event(statusfsm.Stop)
	return &control.StopResponse{}, nil
}

func (s *ControlService) Reload(ctx context.Context, req *control.ReloadRequest) (*control.ReloadResponse, error) {
	if req == nil {
		return nil, common.InvalidArgumentError("request is nil")
	} else if req.ConfigName == "" {
		return nil, common.InvalidArgumentError("missing config name")
	}

	log.Debug("Received reload request", log.StringField("config_name", req.ConfigName))
	s.fsm.Event(statusfsm.Reload, req.ConfigName)
	return &control.ReloadResponse{}, nil
}
