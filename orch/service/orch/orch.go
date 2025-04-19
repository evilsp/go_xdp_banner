package orch

import (
	"context"

	"xdp-banner/api/orch/v1/orch"
	logic "xdp-banner/orch/logic/orch"
	"xdp-banner/orch/service/convert"
	"xdp-banner/pkg/server/common"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

type OrchService struct {
	logic *logic.Orch
	orch.UnimplementedOrchServiceServer
}

func (s *OrchService) RegisterGrpcService(gs grpc.ServiceRegistrar) {
	orch.RegisterOrchServiceServer(gs, s)
}

func (s *OrchService) PublicGrpcMethods() []string {
	return nil
}

func (s *OrchService) RegisterHttpService(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return orch.RegisterOrchServiceHandler(ctx, mux, conn)
}

func NewOrchService(orchLogic *logic.Orch) *OrchService {
	return &OrchService{logic: orchLogic}
}

func (s *OrchService) GetInfo(ctx context.Context, r *orch.GetInfoRequest) (*orch.GetInfoResponse, error) {
	if r.Name == "" {
		return nil, common.InvalidArgumentError("name is required")
	}

	i, err := s.logic.GetInfo(ctx, r.Name)
	if err != nil {
		return nil, common.HandleError(err)
	}

	dto, err := convert.OrchInfoToDto(i)
	if err != nil {
		return nil, common.HandleError(err)
	}

	return &orch.GetInfoResponse{Info: dto}, nil
}

func (s *OrchService) ListInfo(ctx context.Context, r *orch.ListInfoRequest) (*orch.ListInfoResponse, error) {
	if r.Pagesize <= 0 {
		return nil, common.InvalidArgumentError("page size must be greater than 0")
	}

	li, err := s.logic.ListInfo(ctx, r.Pagesize, r.Cursor)
	if err != nil {
		return nil, common.HandleError(err)
	}

	dto, err := convert.OrchInfoItemsToDto(li.Items)
	if err != nil {
		return nil, common.HandleError(err)
	}

	return &orch.ListInfoResponse{
		Total:       li.TotalCount,
		TotalPage:   li.TotalPage,
		CurrentPage: li.CurrentPage,
		HasNext:     li.HasNext,
		NextCursor:  li.NextCursor,
		Items:       dto,
	}, nil
}
