package strategy

import (
	"context"
	api "xdp-banner/api/orch/v1/strategy"

	"xdp-banner/orch/logic/strategy"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

type StrategyService struct {
	api.UnimplementedStrategyServiceServer

	sl *strategy.Strategy
	al *strategy.Applied
}

func New(sl *strategy.Strategy, al *strategy.Applied) *StrategyService {
	return &StrategyService{
		sl: sl,
		al: al,
	}
}

func (s *StrategyService) RegisterGrpcService(gs grpc.ServiceRegistrar) {
	api.RegisterStrategyServiceServer(gs, s)
}

func (s *StrategyService) PublicGrpcMethods() []string {
	return nil
}

func (s *StrategyService) RegisterHttpService(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return api.RegisterStrategyServiceHandler(ctx, mux, conn)
}
