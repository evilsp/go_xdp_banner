package control

import (
	"context"
	"fmt"
	"net"
	"xdp-banner/api/orch/v1/agent/control"
	logic "xdp-banner/orch/logic/agent/control"
	"xdp-banner/orch/service/convert"
	"xdp-banner/pkg/server/common"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"
)

type ControlService struct {
	logic *logic.Control
	control.UnimplementedControlServiceServer
}

func NewControlService(logic *logic.Control) *ControlService {
	return &ControlService{
		logic: logic,
	}
}

func (s *ControlService) RegisterGrpcService(gs grpc.ServiceRegistrar) {
	control.RegisterControlServiceServer(gs, s)
}

func (s *ControlService) PublicGrpcMethods() []string {
	return []string{
		control.ControlService_Init_FullMethodName,
	}
}

func (s *ControlService) RegisterHttpService(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return control.RegisterControlServiceHandler(ctx, mux, conn)
}

func (s *ControlService) Init(ctx context.Context, r *control.InitRequest) (*control.InitResponse, error) {
	switch {
	case r.Name == "":
		return nil, common.InvalidArgumentError("name is required")
	case r.Token == "":
		return nil, common.InvalidArgumentError("token is required")
	case r.PubKeyPem == nil:
		return nil, common.InvalidArgumentError("pub_key_pem is required")
	}

	ipAddress := make([]net.IP, 0, len(r.IpAddresses))
	for _, ipStr := range r.IpAddresses {
		ip := net.ParseIP(ipStr)
		if ip == nil {
			return nil, common.InvalidArgumentError("invalid ip address")
		}
		ipAddress = append(ipAddress, ip)
	}
	fmt.Println("init is starting")
	cert, ca, err := s.logic.Init(ctx, r.Name, r.Token, ipAddress, r.PubKeyPem)
	if err != nil {
		return nil, common.HandleError(err)
	}

	return &control.InitResponse{Cert: cert, Ca: ca}, nil
}

func (s *ControlService) Register(ctx context.Context, r *control.RegisterRequest) (*control.RegisterResponse, error) {
	switch {
	case r.Name == "":
		return nil, common.InvalidArgumentError("name is required")
	}

	token, err := s.logic.RegisterNode(ctx, r.Name)
	if err != nil {
		return nil, common.HandleError(err)
	}

	return &control.RegisterResponse{
		Token: token,
	}, nil
}

func (s *ControlService) Unregister(ctx context.Context, r *control.UnRegisterRequest) (*control.UnRegisterResponse, error) {
	switch {
	case r.Name == "":
		return nil, common.InvalidArgumentError("name is required")
	}

	err := s.logic.UnRegisterNode(ctx, r.Name)
	if err != nil {
		return nil, common.HandleError(err)
	}

	return nil, nil
}

func (s *ControlService) ListRegistration(ctx context.Context, r *control.ListRegistrationRequest) (*control.ListRegistrationResponse, error) {
	switch {
	case r == nil:
		return nil, common.InvalidArgumentError("request is required")
	case r.Pagesize <= 0:
		return nil, common.InvalidArgumentError("pagesize must be greater than 0")
	}

	registrations, err := s.logic.ListRegistration(ctx, r.Pagesize, r.Cursor)
	if err != nil {
		return nil, common.HandleError(err)
	}

	dto, err := convert.RegistrationItemsToDto(registrations.Items)
	if err != nil {
		return nil, common.HandleError(err)
	}

	return &control.ListRegistrationResponse{
		Total:        registrations.TotalCount,
		TotalPage:    registrations.TotalPage,
		HasNext:      registrations.HasNext,
		NextCursor:   registrations.NextCursor,
		Registration: dto,
	}, nil
}

func (s *ControlService) Enable(ctx context.Context, r *control.EnableRequest) (*control.EnableResponse, error) {
	switch {
	case r.Name == "":
		return nil, common.InvalidArgumentError("name is required")
	}

	err := s.logic.Enable(ctx, r.Name, r.Enable)
	if err != nil {
		return nil, common.HandleError(err)
	}

	return nil, nil
}

func (s *ControlService) SetConfig(ctx context.Context, r *control.SetConfigRequest) (*control.SetConfigResponse, error) {
	switch {
	case r.Name == "":
		return nil, common.InvalidArgumentError("name is required")
	case r.ConfigName == "":
		return nil, common.InvalidArgumentError("config is required")
	}

	err := s.logic.SetConfig(ctx, r.Name, r.ConfigName)
	if err != nil {
		return nil, common.HandleError(err)
	}

	return nil, nil
}

func (s *ControlService) GetConfig(ctx context.Context, r *control.GetConfigRequest) (*control.GetConfigResponse, error) {
	switch {
	case r.Name == "":
		return nil, common.InvalidArgumentError("name is required")
	}

	configName, err := s.logic.GetConfig(ctx, r.Name)
	if err != nil {
		return nil, common.HandleError(err)
	}

	return &control.GetConfigResponse{
		ConfigName: configName,
	}, nil
}

func (s *ControlService) SetLabels(ctx context.Context, r *control.SetLabelsRequest) (*control.SetLabelsResponse, error) {
	switch {
	case r.Name == "":
		return nil, common.InvalidArgumentError("name is required")
	}

	err := s.logic.SetLabels(ctx, r.Name, r.Labels)
	if err != nil {
		return nil, common.HandleError(err)
	}

	return nil, nil
}

func (s *ControlService) GetLabels(ctx context.Context, r *control.GetLabelsRequest) (*control.GetLabelsResponse, error) {
	switch {
	case r.Name == "":
		return nil, common.InvalidArgumentError("name is required")
	}

	labels, err := s.logic.GetLabels(ctx, r.Name)
	if err != nil {
		return nil, common.HandleError(err)
	}

	return &control.GetLabelsResponse{
		Labels: labels,
	}, nil
}

func (s *ControlService) GetInfo(ctx context.Context, r *control.GetInfoRequest) (*control.GetInfoResponse, error) {
	switch {
	case r.Name == "":
		return nil, common.InvalidArgumentError("name is required")
	}

	info, err := s.logic.GetInfo(ctx, r.Name)
	if err != nil {
		return nil, common.HandleError(err)
	}

	dto, err := convert.AgentInfoToDto(info)
	if err != nil {
		return nil, common.HandleError(err)
	}

	return &control.GetInfoResponse{
		Info: dto,
	}, nil
}

func (s *ControlService) GetStatus(ctx context.Context, r *control.GetStatusRequest) (*control.GetStatusResponse, error) {
	switch {
	case r.Name == "":
		return nil, common.InvalidArgumentError("name is required")
	}

	status, err := s.logic.GetStatus(ctx, r.Name)
	if err != nil {
		return nil, common.HandleError(err)
	}

	dto, err := convert.AgentStatusToDto(status)
	if err != nil {
		return nil, common.HandleError(err)
	}

	return &control.GetStatusResponse{
		Status: dto,
	}, nil
}

func (s *ControlService) GetAgent(ctx context.Context, r *control.GetAgentRequest) (*control.GetAgentResponse, error) {
	switch {
	case r.Name == "":
		return nil, common.InvalidArgumentError("name is required")
	}

	agent, err := s.logic.GetAgent(ctx, r.Name)
	if err != nil {
		return nil, common.HandleError(err)
	}

	var info *structpb.Struct
	if agent.Info != nil {
		info, err = convert.AgentInfoToDto(agent.Info)
		if err != nil {
			return nil, common.HandleError(err)
		}
	}

	var status *structpb.Struct
	if agent.Status != nil {
		status, err = convert.AgentStatusToDto(agent.Status)
		if err != nil {
			return nil, common.HandleError(err)
		}
	}

	return &control.GetAgentResponse{
		Info:   info,
		Status: status,
	}, nil
}

func (s *ControlService) ListAgents(ctx context.Context, r *control.ListAgentsRequest) (*control.ListAgentsResponse, error) {
	switch {
	case r == nil:
		return nil, common.InvalidArgumentError("request is required")
	case r.Pagesize <= 0:
		return nil, common.InvalidArgumentError("pagesize must be greater than 0")
	}

	agents, err := s.logic.ListAgents(ctx, r.Pagesize, r.Cursor)
	if err != nil {
		return nil, common.HandleError(err)
	}

	dto, err := convert.AgentItemsToDto(agents.Items)
	if err != nil {
		return nil, common.HandleError(err)
	}

	return &control.ListAgentsResponse{
		Total:      agents.TotalCount,
		TotalPage:  agents.TotalPage,
		HasNext:    agents.HasNext,
		NextCursor: agents.NextCursor,
		Agents:     dto,
	}, nil
}
