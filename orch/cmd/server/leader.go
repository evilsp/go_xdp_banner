package server

import (
	"context"
	"sync"
	nodeController "xdp-banner/orch/internal/controller/node"
	strategyController "xdp-banner/orch/internal/controller/strategy"
	"xdp-banner/orch/internal/otlp"
	"xdp-banner/orch/logic"
	applieds "xdp-banner/orch/storage/agent/applied"
	nodes "xdp-banner/orch/storage/agent/node"
	"xdp-banner/pkg/election"
	"xdp-banner/pkg/etcd"
	"xdp-banner/pkg/informer"
	"xdp-banner/pkg/log"
	"xdp-banner/pkg/node"
	"xdp-banner/pkg/wait"

	"go.opentelemetry.io/otel/metric"
)

func runController(ctx context.Context, client etcd.Client, logic *logic.Logic, wg *sync.WaitGroup, errChan chan error) {
	wg.Add(1)
	defer wg.Done()

	nodeName, err := node.Name()
	if err != nil {
		errChan <- err
	}

	isLeader, err := otlp.Meter().Int64Gauge("is_Leader", metric.WithDescription("current node is leader or not"))
	if err != nil {
		errChan <- err
	}
	isLeader.Record(ctx, 0)

	elec, err := election.New(ctx, client, election.NodeInfo{Name: nodeName})
	if err != nil {
		errChan <- err
	}

	leaderChan := elec.Subscribe("controller")

	ws := wait.SingleInstance{
		NewInstance: func() wait.Instance {
			return newController(client, logic)
		},
	}

	go elec.Campaign()

out:
	for {
		select {
		case info := <-leaderChan:
			switch info.Key {
			case election.EventBecomeLeader:
				log.Info("become leader")
				isLeader.Record(ctx, 1)
				ws.Run()
			case election.EventLoseLeader:
				log.Info("lose leader")
				isLeader.Record(ctx, 0)
				ws.Stop()
			case election.EventLeaderChanged:
				log.Info("leader changed")
			default:
				log.Error("unknown leader event", log.StringField("event", info.Key))
			}
		case <-ctx.Done():
			log.Info("exiting controller")
			break out
		}
	}

	elec.Unsubscribe("controller")
	err = elec.Resign(false)
	if err != nil {
		log.Warn("when stop controller, resign leader meet error. we will ignore it, keeping shutdown", log.ErrorField(err))
	}
	ws.Stop()
	elec.StopCampaign()
}

type controller struct {
	ctx        context.Context
	cancel     context.CancelFunc
	wg         wait.Group
	deltaFIFO  *informer.DeltaFIFO
	reflectors []*informer.Reflector
	informer   informer.Informer
	nc         nodeController.NodeController
	sc         strategyController.StrategyController
}

func newController(client etcd.Client, logic *logic.Logic) *controller {
	deltaFIFO := informer.NewDeltaFIFOWithWait(2)

	reflectors := []*informer.Reflector{
		informer.NewReflector(client, "agentnode_reflector", nodes.EtcdDir, deltaFIFO),
		informer.NewReflector(client, "applied_reflector", applieds.EtcdDirRunning, deltaFIFO),
	}
	i := informer.New(deltaFIFO)

	nc := nodeController.New(i)
	sc := strategyController.New(i, logic.Control, logic.Applied)

	ctx, cancel := context.WithCancel(context.Background())

	return &controller{
		ctx:        ctx,
		cancel:     cancel,
		deltaFIFO:  deltaFIFO,
		reflectors: reflectors,
		informer:   i,
		nc:         nc,
		sc:         sc,
	}
}

func (c *controller) Run() {
	for _, r := range c.reflectors {
		c.wg.StartWithContext(c.ctx, r.Run)
	}

	c.wg.StartWithContext(c.ctx, c.informer.Run)
	c.wg.StartWithContext(c.ctx, c.nc.Run)
	c.wg.StartWithContext(c.ctx, c.sc.Run)
}

func (c *controller) Stop() {
	c.cancel()
	c.wg.Wait()
}
