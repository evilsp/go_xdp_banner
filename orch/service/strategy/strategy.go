package strategy

import (
	"context"

	api "xdp-banner/api/orch/v1/strategy"
	"xdp-banner/orch/service/convert"
	"xdp-banner/pkg/server/common"

	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *StrategyService) AddStrategy(ctx context.Context, req *api.Strategy) (*emptypb.Empty, error) {
	sm, err := convert.StrategyDtoToModel(req)
	if err != nil {
		return nil, err
	}

	err = s.sl.Add(ctx, sm)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (s *StrategyService) DeleteStrategy(ctx context.Context, req *api.DeleteStrategyRequest) (*emptypb.Empty, error) {
	err := s.sl.Delete(ctx, req.Name)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (s *StrategyService) GetStrategy(ctx context.Context, req *api.GetStrategyRequest) (*api.Strategy, error) {
	sm, err := s.sl.Get(ctx, req.Name)
	if err != nil {
		return nil, err
	}

	dto, err := convert.StrategyModelToDto(sm)
	if err != nil {
		return nil, err
	}

	return dto, nil
}

func (s *StrategyService) UpdateStrategy(ctx context.Context, req *api.Strategy) (*emptypb.Empty, error) {
	sm, err := convert.StrategyDtoToModel(req)
	if err != nil {
		return nil, err
	}

	err = s.sl.Update(ctx, sm)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (s *StrategyService) ListStrategy(ctx context.Context, req *api.ListStrategyRequest) (*api.ListStrategyResponse, error) {
	strategies, err := s.sl.List(ctx, req.Pagesize, req.Cursor)
	if err != nil {
		return nil, err
	}

	dtos, err := convert.StrategyListItemToDto(strategies.Items)
	if err != nil {
		return nil, err
	}

	return &api.ListStrategyResponse{
		Total:       strategies.TotalCount,
		TotalPage:   strategies.TotalPage,
		CurrentPage: strategies.CurrentPage,
		NextCursor:  strategies.NextCursor,
		HasNext:     strategies.HasNext,
		Items:       dtos,
	}, nil
}

func (s *StrategyService) ApplyStrategy(ctx context.Context, req *api.ApplyStrategyRequest) (*emptypb.Empty, error) {
	if req.Strategy == "" {
		return nil, common.InvalidArgumentError("name is required")
	}

	err := s.al.Create(ctx, req.Strategy)

	return nil, err
}
