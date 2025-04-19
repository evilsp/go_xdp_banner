package rule

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc32"
	"strconv"
	"strings"
	"sync"
	"time"
	"xdp-banner/orch/model/common"
	model "xdp-banner/orch/model/rule"
	"xdp-banner/orch/storage/agent"
	"xdp-banner/pkg/etcd"
	"xdp-banner/pkg/rule"

	"github.com/google/uuid"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	EtcdDir      etcd.Key = etcd.Join(agent.EtcdDir, "rule/")
	EtcdNamesDir etcd.Key = etcd.Join(agent.EtcdDir, "ruleNames/")
)

var (
	// ErrRuleNotFound is returned when the rule is not found.
	ErrRuleNotFound = errors.New("rule not found")
	// ErrRuleAlreadyExists is returned when the rule already exists.
	ErrRuleAlreadyExists = errors.New("rule already exists")
)

var (
	operationLock sync.Mutex
)

// Return rulename + actual rule path
func RuleKey(name string, ruleInfo string) etcd.Key {
	return etcd.Join(EtcdDir, etcd.Join(name, ruleInfo))
}

// Storage 提供对 etcd 的操作
type Storage struct {
	client etcd.Client
}

func New(client etcd.Client) Storage {
	return Storage{client: client}
}

// Add creates a new rule in etcd.
// If the rule already exists, it will return an error.
func (s Storage) Add(ctx context.Context, name string, rule *model.Rule) error {

	operationLock.Lock()
	defer operationLock.Unlock()
	key := RuleKey(name, rule.RuleInfo.Key())
	identityKey := RuleKey(name, rule.RuleInfo.IdentityKey())

	// 获取 ruleNames 列表
	var ruleNames []string
	resp, err := s.client.Get(ctx, EtcdNamesDir)
	if err == etcd.ErrKeyNotFound {
		fmt.Printf("Initialize ruleNames")
	}
	if err != nil && err != etcd.ErrKeyNotFound {
		return fmt.Errorf("get ruleNames went error: %w", err)
	}
	if len(resp.Kvs) > 0 {
		if err := json.Unmarshal(resp.Kvs[0].Value, &ruleNames); err != nil {
			return fmt.Errorf("failed to unmarshal rule names: %w", err)
		}
	} else {
		ruleNames = []string{}
		fmt.Println("RuleName list initialized")
	}

	// 检查 name 是否已存在
	nameExists := false
	for _, existing := range ruleNames {
		if existing == name {
			nameExists = true
			break
		}
	}
	// 如果不存在，则添加进去
	if !nameExists {
		ruleNames = append(ruleNames, name)
	}
	updatedNamesValue, err := json.Marshal(ruleNames)
	if err != nil {
		return fmt.Errorf("failed to marshal updated rule names: %w", err)
	}

	// 获取规则过期时间对应的 lease TTL（要求 expiresAt 在未来）
	ttl := time.Until(rule.RuleMeta.ExpiresAt)
	if ttl <= 0 {
		return fmt.Errorf("ExpiresAt must be in the future")
	}

	leaseResp, err := s.client.Grant(ctx, ttl)
	if err != nil {
		return fmt.Errorf("failed to create lease: %w", err)
	}

	// 首先检查 identityKey 是否已存在
	idResp, err := s.client.Get(ctx, identityKey)
	if err != nil {
		return fmt.Errorf("failed to get identityKey: %w", err)
	}
	if len(idResp.Kvs) > 0 && string(idResp.Kvs[0].Value) != "0" {
		// identityKey 已存在，并且不是 0 值，则直接复用该 identity
		existingIdentity := string(idResp.Kvs[0].Value)
		rule.RuleMeta.Identity = existingIdentity

		// 使用一个事务写入规则，判断条件仅为 key 不存在
		txn, cancel := s.client.Txn(ctx)
		defer cancel()
		if nameExists {
			// name 已存在，直接写规则
			txn = txn.If(
				clientv3.Compare(clientv3.Version(key), "=", 0),
			).Then(
				clientv3.OpPut(key, rule.RuleMeta.MarshalStr(), clientv3.WithLease(leaseResp.ID)),
			)
		} else {
			// name 不存在，除了写规则，还需要更新 EtcdNamesDir
			txn = txn.If(
				clientv3.Compare(clientv3.Version(key), "=", 0),
			).Then(
				clientv3.OpPut(key, rule.RuleMeta.MarshalStr(), clientv3.WithLease(leaseResp.ID)),
				clientv3.OpPut(EtcdNamesDir, string(updatedNamesValue)),
			)
		}
		txnResp, err := txn.Commit()
		if err != nil {
			s.client.Revoke(ctx, leaseResp.ID)
			return fmt.Errorf("etcd transaction failed: %w", err)
		}
		if !txnResp.Succeeded {
			return ErrRuleAlreadyExists
		}
		return nil
	}

	// 如果 identityKey 不存在
	u := uuid.New()
	// 2. 对 UUID 的 16 字节做 CRC32
	sum := crc32.ChecksumIEEE(u[:])
	// 3. 转为十进制字符串
	newIdentity := strconv.FormatUint(uint64(sum), 10)
	rule.RuleMeta.Identity = newIdentity

	txn, cancel := s.client.Txn(ctx)
	defer cancel()
	txn = txn.If(
		clientv3.Compare(clientv3.Version(key), "=", 0),
		clientv3.Compare(clientv3.Version(identityKey), "=", 0),
	).Then(
		clientv3.OpPut(key, rule.RuleMeta.MarshalStr(), clientv3.WithLease(leaseResp.ID)),
		clientv3.OpPut(EtcdNamesDir, string(updatedNamesValue)),
		clientv3.OpPut(identityKey, newIdentity, clientv3.WithLease(leaseResp.ID)),
	)
	txnResp, err := txn.Commit()
	if err != nil {
		s.client.Revoke(ctx, leaseResp.ID)
		return fmt.Errorf("etcd transaction failed: %w", err)
	}
	if !txnResp.Succeeded {
		return ErrRuleAlreadyExists
	}

	return nil
}

// Delete deletes a config from etcd.
func (s Storage) Delete(ctx context.Context, name string, rule *model.Rule) error {
	operationLock.Lock()
	defer operationLock.Unlock()
	key := RuleKey(name, rule.RuleInfo.Key())
	_, err := s.client.Delete(ctx, key)
	return err
}

// Update updates a config in etcd.
func (s Storage) Update(ctx context.Context, name string, rule *model.Rule) error {
	operationLock.Lock()
	defer operationLock.Unlock()
	key := RuleKey(name, rule.RuleInfo.Key())
	identityKey := RuleKey(name, rule.RuleInfo.IdentityKey())

	idResp, err := s.client.Get(ctx, identityKey)
	if err != nil {
		return fmt.Errorf("failed to get identityKey: %w", err)
	}

	if len(idResp.Kvs) > 0 {
		rule.RuleMeta.Identity = string(idResp.Kvs[0].Value)
		err = s.client.Update(ctx, key, rule.RuleMeta.MarshalStr())
		if err != nil {
			if err == etcd.ErrKeyNotFound {
				return ErrRuleNotFound
			}
			return err
		}
		return nil
	}

	return fmt.Errorf("IdentityKey Generation went error")
}

// GetConfig gets a config from etcd.
func (s Storage) GetRule(ctx context.Context, name string) (model.RuleItem, error) {

	prefix := RuleKey(name, "")

	resp, err := s.client.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("failed to get rule from etcd: %w", err)
	}

	if len(resp.Kvs) == 0 {
		return nil, ErrRuleNotFound
	}

	ruleList := make(model.RuleItem, 0, len(resp.Kvs))

	for _, kv := range resp.Kvs {
		relPath := strings.TrimPrefix(string(kv.Key), prefix)

		var info rule.RuleInfo
		var meta rule.RuleMeta

		parts := strings.Split(strings.Trim(relPath, "/"), "/")
		if len(parts) != 3 {
			continue
		}

		sportDport := strings.Split(parts[2], "-")
		if len(sportDport) != 2 {
			continue
		}

		sport, _ := strconv.Atoi(sportDport[0])
		dport, _ := strconv.Atoi(sportDport[1])

		info = rule.RuleInfo{
			Cidr:     parts[0],
			Protocol: parts[1],
			Sport:    uint16(sport),
			Dport:    uint16(dport),
		}

		if err := json.Unmarshal(kv.Value, &meta); err != nil {
			return nil, fmt.Errorf("failed to unmarshal rule meta: %w", err)
		}

		ruleList = append(ruleList, model.Rule{
			RuleInfo: info,
			RuleMeta: meta,
		})
	}

	if len(ruleList) == 0 {
		return nil, ErrRuleNotFound
	}

	return ruleList, nil
}

// List Name by add
func (s Storage) List(ctx context.Context, size int64, nextCursor string) (*model.RuleList, error) {

	// Get existing names from etcd
	resp, err := s.client.Get(ctx, EtcdNamesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get key from etcd: %w", err)
	}

	if len(resp.Kvs) != 1 {
		return nil, fmt.Errorf("expected exactly one key-value pair, got %d", len(resp.Kvs))
	}

	kv := resp.Kvs[0]
	if len(kv.Value) == 0 {
		return nil, errors.New("empty value in etcd response")
	}

	var ruleNames []string
	if err := json.Unmarshal(kv.Value, &ruleNames); err != nil {
		return nil, fmt.Errorf("failed to unmarshal etcd value: %w", err)
	}

	if len(ruleNames) == 0 {
		return nil, errors.New("no names found in etcd data")
	}

	start := 0
	if nextCursor != "" {
		for i, name := range ruleNames {
			if name == nextCursor {
				start = i
				break
			}
		}
	}

	end := start + int(size)
	if end > len(ruleNames) {
		end = len(ruleNames)
	}

	selectNames := ruleNames[start:end]

	var wg sync.WaitGroup
	result := make(model.RuleItems)
	var resultLock sync.Mutex
	var firstErr error

	sem := make(chan struct{}, 10) // limit to 10 concurrent goroutines

	for _, ruleName := range selectNames {
		wg.Add(1)
		sem <- struct{}{} // acquire semaphore
		go func(p string) {
			defer func() {
				<-sem // release semaphore
				wg.Done()
			}()
			val, err := s.GetRule(ctx, p)
			if err != nil {
				resultLock.Lock()
				if firstErr == nil {
					firstErr = fmt.Errorf("get prefix %s failed: %w", p, err)
				}
				resultLock.Unlock()
				return
			}
			resultLock.Lock()
			result[p] = val
			resultLock.Unlock()
		}(etcd.Join(EtcdDir, ruleName))
	}
	wg.Wait()
	close(sem)

	if firstErr != nil {
		return nil, firstErr
	}

	hasNext := end < len(ruleNames)
	next := ""
	if hasNext {
		next = ruleNames[end]
	}

	return &model.RuleList{
		List: common.List{
			TotalCount: int64(len(ruleNames)),
			HasNext:    hasNext,
			NextCursor: next,
		},
		Items: result,
	}, nil
}

func (s Storage) DeleteDir(ctx context.Context) error {
	operationLock.Lock()
	defer operationLock.Unlock()
	return s.client.DeleteWithPrefix(ctx, EtcdDir)
}
