package rule

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"xdp-banner/api/orch/v1/rule"
	"xdp-banner/pkg/etcd"
	"xdp-banner/pkg/log"
	model "xdp-banner/pkg/rule"

	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

var ErrValItem = errors.New("rule val from refletor not string")
var ErrSkipItem = errors.New("skip this item")

func convertMapToWatchRuleResponses(
	itemsMap map[string]any,
	prefix string,
) ([]*rule.WatchRuleResponse, error) {

	var results []*rule.WatchRuleResponse

	for key, val := range itemsMap {

		if !hasPrefix(key, prefix) {
			continue
		}

		ruleVal, err := parseRuleMeta(val)
		switch {
		case err == ErrSkipItem:
			continue
		case err == ErrValItem:
			log.Warn("invalid rule val, skip", zap.String("key", key))
			continue
		case err != nil:
			return nil, fmt.Errorf("parseRuleMeta failed for key %q: %w", key, err)
		}

		valForGrpc, err := convertToStruct(*ruleVal)
		if err != nil {
			log.Error("event onAdd process went error", zap.Error(err))
		}

		// 构造 WatchRuleResponse，对初次列出的数据可统一视为 “ADD” 或其它类型
		results = append(results, &rule.WatchRuleResponse{
			RuleKey:   key,
			RuleVal:   valForGrpc,
			EventType: 0,
		})
		log.Info("Send Rule key", zap.String("key", key))
	}
	return results, nil
}

func parseRuleMeta(val any) (*model.RuleMeta, error) {

	raw, ok := val.(string)
	if !ok {
		return nil, ErrValItem
	}

	// Parse key like  "/agent/rule/default/2001:da8:c807:20::/64/TCP/0-22"
	// Parse Val like "{\"comment\":\"Example IPv6 block\",\"created_at\":\"2025-04-19T05:22:37.555547563Z\",\"expires_at\":\"2025-04-19T06:22:37.555547563Z\",\"identity\":\"3009407147\"}"

	// 反序列化为 model.RuleMeta
	var ruleVal model.RuleMeta
	if err := json.Unmarshal([]byte(raw), &ruleVal); err == nil {
		log.Info("Send Rule val", zap.String("val", raw))
		return &ruleVal, nil
	}

	// Check if identity kv
	if len(raw) == 10 {
		if _, err2 := strconv.ParseInt(raw, 10, 64); err2 == nil {
			// 确实是 10 位数字，就静默跳过
			return nil, ErrSkipItem
		}
	}

	return nil, ErrValItem
}

// convertToStruct 将任意对象转换为 google.protobuf.Struct 类型
func convertToStruct(obj any) (*structpb.Struct, error) {
	b, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return structpb.NewStruct(m)
}

// tempEventHandler 用于接收 informer 下发的事件，并写入 eventCh
type tempEventHandler struct {
	prefix string
	ch     chan *rule.WatchRuleResponse
}

func (h *tempEventHandler) OnAdd(key etcd.Key, obj any, isInInitialList bool) {

	log.Info("Processing key", zap.String("key", key))

	if !hasPrefix(string(key), h.prefix) {
		return
	}

	ruleVal, err := parseRuleMeta(obj)
	switch {
	case err == ErrSkipItem:
		return
	case err == ErrValItem:
		log.Warn("invalid rule val, skip", zap.String("key", key))
		return
	case err != nil:
		log.Error("event onAdd process went error", zap.Error(err))
		return
	}

	valForGrpc, err := convertToStruct(*ruleVal)
	if err != nil {
		log.Error("event onAdd process went error", zap.Error(err))
	}

	h.ch <- &rule.WatchRuleResponse{
		RuleKey:   string(key),
		RuleVal:   valForGrpc,
		EventType: 0,
	}

	log.Info("Send Rule key", zap.String("key", key))
}

func (h *tempEventHandler) OnUpdate(key etcd.Key, oldObj, newObj any) {
	if !hasPrefix(string(key), h.prefix) {
		return
	}

	ruleVal, err := parseRuleMeta(newObj)
	switch {
	case err == ErrSkipItem:
		return
	case err == ErrValItem:
		log.Warn("invalid rule val, skip", zap.String("key", key))
		return
	case err != nil:
		log.Error("event onUpdate process went error", zap.Error(err))
		return
	}

	valForGrpc, err := convertToStruct(*ruleVal)
	if err != nil {
		log.Error("event onUpdate process went error", zap.Error(err))
	}

	h.ch <- &rule.WatchRuleResponse{
		RuleKey:   string(key),
		RuleVal:   valForGrpc,
		EventType: 0,
	}

	log.Info("Send Rule key", zap.String("key", key))
}

func (h *tempEventHandler) OnDelete(key etcd.Key, obj any) {
	if !hasPrefix(string(key), h.prefix) {
		return
	}

	ruleVal, err := parseRuleMeta(obj)
	switch {
	case err == ErrSkipItem:
		return
	case err == ErrValItem:
		log.Warn("invalid rule val, skip", zap.String("key", key))
		return
	case err != nil:
		log.Error("event onDelete process went error", zap.Error(err))
		return
	}

	valForGrpc, err := convertToStruct(*ruleVal)
	if err != nil {
		log.Error("event onDelete process went error", zap.Error(err))
	}

	h.ch <- &rule.WatchRuleResponse{
		RuleKey:   string(key),
		RuleVal:   valForGrpc,
		EventType: 1,
	}

	log.Info("Delete Rule key", zap.String("key", key))
}

// hasPrefix 简单判断 key 是否以 prefix 开头
func hasPrefix(key, prefix string) bool {
	return len(key) >= len(prefix) && key[:len(prefix)] == prefix
}
