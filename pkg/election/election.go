package election

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"xdp-banner/pkg/etcd"
	"xdp-banner/pkg/log"
	"xdp-banner/pkg/pubsub"

	"go.etcd.io/etcd/client/v3/concurrency"
)

type Election struct {
	*pubsub.PubSub
	ctx    context.Context
	cancel context.CancelFunc

	mutex        sync.Mutex
	isCampaining bool
	cancelCamp   context.CancelFunc
	isLeader     bool

	opt      *Options
	session  *concurrency.Session
	election *concurrency.Election
}

type NodeInfo struct {
	Name       string
	ListenAddr string
}

func (i NodeInfo) Marshal() string {
	s, _ := json.Marshal(i)
	return string(s)
}

func (i *NodeInfo) Unmarshal(s string) error {
	return json.Unmarshal([]byte(s), i)
}

func New(ctx context.Context, client etcd.Client, ni NodeInfo, opt ...Option) (*Election, error) {
	opt = append(opt, WithLeaderVal(ni.Marshal()))
	options := ParseOptions(opt...)

	s, err := concurrency.NewSession(client.RawClient(), concurrency.WithTTL(options.TTL))
	if err != nil {
		return nil, fmt.Errorf("failed to create etcd session: %w", err)
	}

	c, cancel := context.WithCancel(ctx)

	ee := concurrency.NewElection(s, options.Prefix)

	e := &Election{
		PubSub: pubsub.New(),
		ctx:    c,
		cancel: cancel,

		opt:      options,
		session:  s,
		election: ee,
	}

	e.runObserver()

	return e, nil
}

func (e *Election) runObserver() {
	ctx := e.ctx
	event := e.election.Observe(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return

			case ev := <-event:
				if ev.Kvs == nil && len(ev.Kvs) == 0 {
					log.Debug("unexpected event without kvs", log.AnyField("event", ev))
					continue
				}

				// Skip if the leader is itself, Campaign() will handle it
				if string(ev.Kvs[0].Value) == e.opt.LeaderVal {
					continue
				}

				ni := NodeInfo{}
				if err := ni.Unmarshal(string(ev.Kvs[0].Value)); err != nil {
					log.Error("unmarshal leader value failed", log.ErrorField(err))
					continue
				}

				e.Publish(EventLeaderChanged, ni)
			}
		}
	}()
}

func (e *Election) Campaign() {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if e.isLeader || e.isCampaining {
		return
	}
	e.isCampaining = true

	ctx, cancel := context.WithCancel(e.ctx)
	e.cancelCamp = cancel

	afterCamp := func() {
		e.mutex.Lock()
		defer e.mutex.Unlock()

		e.isLeader = true
		e.isCampaining = false
		e.Publish(EventBecomeLeader, nil)
	}

	go func() {
		defer afterCamp()

		if err := e.election.Campaign(ctx, e.opt.LeaderVal); err != nil {
			log.Error("fail to campaign", log.ErrorField(err))
		}
	}()
}

func (e *Election) StopCampaign() {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if !e.isCampaining {
		return
	}
	e.isCampaining = false

	if e.cancelCamp != nil {
		e.cancelCamp()
	}

	log.Info("stop campaign")
}

// Resign resigns the leadership. If force is true, it will emit resign event even if it's not a leader.
func (e *Election) Resign(force bool) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if !force && !e.isLeader {
		return nil
	}

	if err := e.election.Resign(e.ctx); err != nil {
		return err
	}

	e.isLeader = false
	e.Publish(EventLoseLeader, nil)

	return nil
}

func (e *Election) Close() {
	e.cancel()
	e.session.Close()
}

func (e *Election) Leader(ctx context.Context) (ni NodeInfo, err error) {
	resp, err := e.election.Leader(ctx)
	if err != nil {
		return ni, fmt.Errorf("failed to get leader: %w", err)
	}

	if err := ni.Unmarshal(string(resp.Kvs[0].Value)); err != nil {
		return ni, fmt.Errorf("failed to unmarshal leader value: %w", err)
	}

	return ni, nil
}
