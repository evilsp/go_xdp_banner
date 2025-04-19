package strategy

import (
	"context"
	api "xdp-banner/api/orch/v1/strategy"
	"xdp-banner/orch/service/convert"
	"xdp-banner/pkg/errors"

	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *StrategyService) GetRunningStrategy(ctx context.Context, req *api.GetRunningAppliedRequest) (*api.Applied, error) {
	if req.Name == "" {
		return nil, errors.NewInputError("name is required")
	}

	applied, err := s.al.GetRunning(ctx, req.Name)
	if err != nil {
		return nil, err
	}

	dto, err := convert.AppliedModelToDto(applied)
	if err != nil {
		return nil, err
	}

	return dto, nil
}

func (s *StrategyService) ListRunningApplied(ctx context.Context, req *api.ListRunningAppliedRequest) (*api.ListRunningAppliedResponse, error) {
	applieds, err := s.al.ListRunning(ctx, req.Pagesize, req.Cursor)
	if err != nil {
		return nil, err
	}

	dtos, err := convert.AppliedListItemToDto(applieds.Items)
	if err != nil {
		return nil, err
	}

	return &api.ListRunningAppliedResponse{
		Total:       applieds.TotalCount,
		TotalPage:   applieds.TotalPage,
		CurrentPage: applieds.CurrentPage,
		HasNext:     applieds.HasNext,
		NextCursor:  applieds.NextCursor,
		Items:       dtos,
	}, nil
}

func (s *StrategyService) GetHistoryApplied(ctx context.Context, req *api.GetHistoryAppliedRequest) (*api.Applied, error) {
	if req.Name == "" {
		return nil, errors.NewInputError("name is required")
	}

	applied, err := s.al.GetHistory(ctx, req.Name)
	if err != nil {
		return nil, err
	}

	dto, err := convert.AppliedModelToDto(applied)
	if err != nil {
		return nil, err
	}

	return dto, nil
}

func (s *StrategyService) ListHistoryApplied(ctx context.Context, req *api.ListHistoryAppliedRequest) (*api.ListHistoryAppliedResponse, error) {
	applieds, err := s.al.ListHistory(ctx, req.Pagesize, req.Cursor)
	if err != nil {
		return nil, err
	}

	dtos, err := convert.AppliedListItemToDto(applieds.Items)
	if err != nil {
		return nil, err
	}

	return &api.ListHistoryAppliedResponse{
		Total:       applieds.TotalCount,
		TotalPage:   applieds.TotalPage,
		CurrentPage: applieds.CurrentPage,
		HasNext:     applieds.HasNext,
		NextCursor:  applieds.NextCursor,
		Items:       dtos,
	}, nil
}

func (s *StrategyService) DeleteHistoryApplied(ctx context.Context, req *api.DeleteHistoryAppliedRequest) (*emptypb.Empty, error) {
	if req.Name == "" {
		return nil, errors.NewInputError("name is required")
	}

	err := s.al.DeleteHistory(ctx, req.Name)

	return nil, err
}
