package report

import (
	"context"
	"xdp-banner/api/orch/v1/agent/report"
	"xdp-banner/pkg/server/common"

	logic "xdp-banner/orch/logic/agent/report"
	model "xdp-banner/orch/model/node"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

type ReportService struct {
	logic *logic.Report
	report.UnimplementedReportServiceServer
}

func NewReportService(logic *logic.Report) *ReportService {
	return &ReportService{
		logic: logic,
	}
}

func (s *ReportService) RegisterGrpcService(gs grpc.ServiceRegistrar) {
	report.RegisterReportServiceServer(gs, s)
}

func (s *ReportService) PublicGrpcMethods() []string {
	return []string{}
}

func (s *ReportService) RegisterHttpService(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return nil
}

func (s *ReportService) Report(ctx context.Context, status *report.Status) (*report.ReportResponse, error) {
	if status == nil {
		return nil, common.InvalidArgumentError("status is required")
	} else if status.Name == "" {
		return nil, common.InvalidArgumentError("name is required")
	}

	m := &model.AgentStatus{
		GrpcEndpoint: status.GrpcEndpoint,
		Config:       status.ConfigName,
		Phase:        status.Phase.String(),
	}
	if status.Error != nil {
		m.Error = &model.ErrorTime{
			Message: status.Error.Message,
			RetryAt: status.Error.RetryAt.AsTime(),
		}
	}

	err := s.logic.UpdateStatus(ctx, status.Name, m)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
