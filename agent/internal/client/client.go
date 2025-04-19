package client

import (
	"context"
	"fmt"
	"xdp-banner/api/orch/v1/agent/report"
	"xdp-banner/api/orch/v1/rule"
	"xdp-banner/pkg/log"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Client interface {
	Close()
	GetRule(ctx context.Context, name string, ruleChan chan *rule.WatchRuleResponse) error
	Report(ctx context.Context, status *report.Status) error
}

type client struct {
	conn *grpc.ClientConn

	rule   rule.RuleServiceClient
	report report.ReportServiceClient
}

func New(endpoint string, opts ...grpc.DialOption) (Client, error) {
	conn, err := grpc.NewClient(
		endpoint,
		opts...,
	)
	if err != nil {
		return nil, fmt.Errorf("grpc client connect failed: %w", err)
	}

	ruc := rule.NewRuleServiceClient(conn)
	rec := report.NewReportServiceClient(conn)

	return &client{
		conn: conn,

		rule:   ruc,
		report: rec,
	}, nil
}

func (c *client) Close() {
	c.conn.Close()
}

// Send rule update into chan
func (c *client) GetRule(ctx context.Context, name string, ruleChan chan *rule.WatchRuleResponse) error {

	req := &rule.WatchRuleRequest{
		RuleName: name, // 想监听的规则名
	}

	// 调用流式 RPC 接口
	stream, err := c.rule.WatchRuleResources(ctx, req)
	if err != nil {
		log.Error("调用 WatchRuleResources 失败: ", zap.Error(err))
		return err
	}

	// 循环消费流数据
	for {
		// 首先阻塞等待服务端返回消息
		resp, err := stream.Recv()
		if err != nil {
			log.Error("Get rule from server went error: ", zap.Error(err))
			return err
		}

		select {
		case <-ctx.Done():
			fmt.Println("状态切换，退出 Rule Fetch 循环")
			return nil
		case ruleChan <- resp: // <- 直接发送指针
			// 正常发送后可继续下一次循环
		}
	}
}

func (c *client) Report(ctx context.Context, status *report.Status) error {
	if status == nil {
		return fmt.Errorf("status is nil")
	}

	_, err := c.report.Report(ctx, status)
	if err != nil {
		return fmt.Errorf("report status: %w", err)
	}

	return nil
}
