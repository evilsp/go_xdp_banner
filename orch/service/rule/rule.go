package rule

import (
	"context"
	"errors"
	"fmt"

	"time"

	"xdp-banner/api/orch/v1/rule"

	"xdp-banner/orch/cmd/global"
	ruleLogic "xdp-banner/orch/logic/rulecenter"
	"xdp-banner/orch/service/convert"
	ruleStorage "xdp-banner/orch/storage/agent/rule"
	"xdp-banner/pkg/etcd"
	"xdp-banner/pkg/informer"
	"xdp-banner/pkg/log"
	"xdp-banner/pkg/server/common"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"
)

type RuleService struct {
	rl *ruleLogic.RuleCenter
	rule.UnimplementedRuleServiceServer
}

func NewRuleService(rl *ruleLogic.RuleCenter) *RuleService {
	return &RuleService{rl: rl}
}

func (s *RuleService) RegisterGrpcService(gs grpc.ServiceRegistrar) {
	rule.RegisterRuleServiceServer(gs, s)
}

func (s *RuleService) PublicGrpcMethods() []string {
	return nil
}

func (s *RuleService) RegisterHttpService(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return rule.RegisterRuleServiceHandler(ctx, mux, conn)
}

func (s *RuleService) AddRule(ctx context.Context, r *rule.AddRuleRequest) (*rule.AddRuleResponse, error) {
	if r.Name == "" {
		return nil, common.InvalidArgumentError("name is required")
	}

	ruleModel, err := convert.RuleDtoToModel(r.Rule)
	if err != nil {
		return nil, common.HandleError(err)
	}

	err = s.rl.AddRule(ctx, r.Name, ruleModel)
	if err != nil {
		return nil, common.HandleError(err)
	}

	return nil, nil
}

func (s *RuleService) DeleteRule(ctx context.Context, r *rule.DeleteRuleRequest) (*rule.DeleteRuleResponse, error) {
	if r.Name == "" {
		return nil, common.InvalidArgumentError("name is required")
	}

	ruleModel, err := convert.RuleDtoToModel(r.Rule)
	if err != nil {
		return nil, common.HandleError(err)
	}

	err = s.rl.DeleteRule(ctx, r.Name, ruleModel)

	if err != nil {
		return nil, common.HandleError(err)
	}

	return nil, nil
}

func (s *RuleService) UpdateRule(ctx context.Context, r *rule.UpdateRuleRequest) (*rule.UpdateRuleResponse, error) {
	if r.Name == "" {
		return nil, common.InvalidArgumentError("name is required")
	}

	ruleModel, err := convert.RuleDtoToModel(r.Rule)
	if err != nil {
		return nil, common.HandleError(err)
	}

	err = s.rl.UpdateRule(ctx, r.Name, ruleModel)
	if err != nil {
		return nil, common.HandleError(err)
	}

	return nil, nil
}

func (s *RuleService) GetRule(ctx context.Context, r *rule.GetRuleRequest) (*rule.GetRuleResponse, error) {
	if r.Name == "" {
		return nil, common.InvalidArgumentError("name is required")
	}

	ruleItems, err := s.rl.GetRule(ctx, r.Name)
	if err != nil {
		return nil, common.HandleError(err)
	}

	combined := &structpb.Struct{
		Fields: make(map[string]*structpb.Value, len(ruleItems)),
	}

	for _, item := range ruleItems {
		respEntry, err := convert.RuleModelToDto(&item)
		if err != nil {
			return nil, fmt.Errorf("failed to convert rule %s: %w", item.RuleInfo.Key(), err)
		}

		// 使用规则名称作为字段名
		combined.Fields[item.RuleInfo.Key()] = &structpb.Value{
			Kind: &structpb.Value_StructValue{
				StructValue: respEntry,
			},
		}
	}

	return &rule.GetRuleResponse{
		Rule: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"items": {
					Kind: &structpb.Value_StructValue{
						StructValue: combined,
					},
				},
				"count": {
					Kind: &structpb.Value_NumberValue{
						NumberValue: float64(len(ruleItems)),
					},
				},
			},
		},
	}, nil
}

func (s *RuleService) ListRule(ctx context.Context, r *rule.ListRuleRequest) (*rule.ListRuleResponse, error) {
	if r.Pagesize <= 0 {
		return nil, common.InvalidArgumentError("page size must be greater than 0")
	}

	rl, err := s.rl.ListRule(ctx, r.Pagesize, r.Cursor)
	if err != nil {
		return nil, common.HandleError(err)
	}

	dto, err := convert.RuleListToDto(rl.Items)
	if err != nil {
		return nil, common.HandleError(err)
	}

	return &rule.ListRuleResponse{
		Total:       rl.TotalCount,
		TotalPage:   rl.TotalPage,
		CurrentPage: rl.CurrentPage,
		HasNext:     rl.HasNext,
		NextCursor:  rl.NextCursor,
		Items:       dto,
	}, nil
}

// 当 gRPC 服务器为一个流式 RPC 调用时，会自动生成一个 context，但不会将其传入第一个参数，而是通过 stream.Context() 返回这个上下文。
// 因此，从功能上来看，它们具有相同的取消信号和截止时间，能够同步响应客户端断开连接等事件。

func (s *RuleService) WatchRuleResources(req *rule.WatchRuleRequest,
	stream rule.RuleService_WatchRuleResourcesServer) error {

	ctx := stream.Context()
	prefix := etcd.Join(ruleStorage.EtcdDir, req.GetRuleName())
	log.Info("WatchRuleName:", zap.String("prefix", prefix))

	inf, ok := informer.GlobalInformerRegistry.Get(prefix)

	if !ok {
		var err error
		inf, err = informer.CreateInformerForPrefix(ctx, global.Cli, prefix)
		if err != nil {
			return err
		}
		informer.GlobalInformerRegistry.Add(prefix, inf)
	}

	syncCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if !inf.WaitForCacheSync(syncCtx) {
		return errors.New("informer cache sync 超时")
	}

	log.Info("Start A new instance for name:", zap.String("rulename", req.GetRuleName()))
	// 创建用于后续事件通知的 channel
	eventCh := make(chan *rule.WatchRuleResponse, 1000)

	// 构造并注册一个临时事件处理器，将事件写入 eventCh
	handler := &tempEventHandler{
		prefix: prefix,
		ch:     eventCh,
	}

	itemsMap, err := inf.RegisterHandlerAndList([]string{prefix}, handler)
	if err != nil {
		return err
	}

	// --- ADDED: 将返回的 map[string]any 转为 WatchRuleResponse 并初次发送给客户端
	// 这里演示一个简单转换：对每个 key/value 都打包成自定义响应
	// 你可根据业务需求灵活处理，比如过滤掉不是 prefix 的、或解析 value。
	initialEvents, err := convertMapToWatchRuleResponses(itemsMap, prefix)
	if err != nil {
		return err
	}

	log.Info("WatchRuleResources: 队列初始化成功", zap.Int("EventsNum", len(initialEvents)))

	for _, evt := range initialEvents {
		if err := stream.Send(evt); err != nil {
			log.Error("WatchRuleResources: 初次发送失败", zap.Error(err))
			return err
		}
		log.Info("WatchRuleResources: 初始化发送成功")
	}

	// --- 进入循环，持续监听 eventCh，将事件流式返回给客户端，直到连接中断
	for {
		select {
		case <-stream.Context().Done():
			log.Info("WatchRuleResources: 客户端断开")
			informer.GlobalInformerRegistry.Delete(prefix)
			return nil
		case event, ok := <-eventCh:
			if !ok {
				return nil
			}
			if err := stream.Send(event); err != nil {
				log.Error("WatchRuleResources: 发送事件失败", zap.Error(err))
				return err
			}
		}
	}
}
