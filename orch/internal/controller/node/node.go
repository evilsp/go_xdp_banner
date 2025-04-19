package node

import (
	"context"
	"fmt"
	"time"

	"xdp-banner/api/orch/v1/agent/report"
	"xdp-banner/orch/internal/otlp"
	nodem "xdp-banner/orch/model/node"
	nodes "xdp-banner/orch/storage/agent/node"

	"xdp-banner/pkg/controller"
	"xdp-banner/pkg/etcd"
	"xdp-banner/pkg/informer"
	"xdp-banner/pkg/log"

	"go.opentelemetry.io/otel/metric"
)

type node struct {
	info   *nodem.AgentInfo
	status *nodem.AgentStatus
}

// gateway
var _ controller.ControllerImpl = (*nodeController)(nil)

type NodeController struct {
	*controller.Controller
}

func New(i informer.Informer) NodeController {
	nodeController := newNodeController()

	controller := controller.New(i, controller.ControllerOption{
		Name: "node",
		Impl: nodeController,
	})

	return NodeController{
		Controller: controller,
	}

}

type nodeController struct {
	cp      clientPool
	metrics metrics
}

func newNodeController() *nodeController {
	return &nodeController{
		cp:      newClientPool(),
		metrics: newMetrics(),
	}
}

func (n *nodeController) SyncHandler(key string, informer informer.Informer) error {
	/*
		get and check node info
	*/
	infoObj, infoExist, err := informer.Get(nodes.InfoKey(key))
	if err != nil {
		return fmt.Errorf("get node info %s failed %w", key, err)
	} else if !infoExist {
		log.Info("node info not exist, skip it now", log.StringField("node", key))
		return nil
	}

	infoStr, ok := infoObj.(string)
	if !ok {
		return fmt.Errorf("node info %s raw obj is not string", key)
	}
	info := new(nodem.AgentInfo)
	err = info.Unmarshal([]byte(infoStr))
	if err != nil {
		return fmt.Errorf("node info %s unmarshal failed %w, in", key, err)
	}

	/*
		get and check node status
	*/
	statusObj, statusExist, err := informer.Get(nodes.StatusKey(key))
	if err != nil {
		return fmt.Errorf("get node status %s failed %w", key, err)
	} else if !statusExist {
		log.Info("node status not exist, skip it now", log.StringField("node", key))
		return nil
	}
	statusStr, ok := statusObj.(string)
	if !ok {
		return fmt.Errorf("node status %s raw obj is not string", key)
	}
	status := new(nodem.AgentStatus)
	err = status.Unmarshal([]byte(statusStr))
	if err != nil {
		return fmt.Errorf("node status %s unmarshal failed %w", key, err)
	}

	// node info and status
	node := node{
		info:   info,
		status: status,
	}

	n.metrics.SyncTotal.Add(context.Background(), 1)
	err = n.reconcile(node)
	if err != nil {
		n.metrics.SyncError.Add(context.Background(), 1)
		return err
	}
	return nil
}

func (r *nodeController) reconcile(node node) error {
	logger := log.With(log.StringField("node", node.info.Name))
	if node.status == nil {
		logger.Info("node status is nil, maybe node is down, skip it now")
		return nil
	}

	if node.status.Error != nil {
		if node.status.Error.RetryAt.After(time.Now()) {
			logger.Debug("node has error, but not retry yet, skip it now", log.StringField("error", node.status.Error.Message), log.AnyField("retryAt", node.status.Error.RetryAt))
			return fmt.Errorf("need retry")
		}
	}

	if isDisableNode(node) {
		logger.Debug("node is disable, let it stop")
		return r.stopNode(node)
	}

	if isReadyNode(node) || isEnableNode(node) {
		logger.Debug("node is ready, let it start", log.StringField("config", node.info.Config))
		return r.startNode(node)
	}

	if needReloadNode(node) {
		logger.Debug("node need reload", log.StringField("newconfig", node.info.Config), log.StringField("oldconfig", node.status.Config))
		return r.reloadNode(node)
	}

	return nil
}

func isDisableNode(node node) bool {
	return !node.info.Enable
}

func (r *nodeController) stopNode(node node) error {
	if node.status.Phase != report.Phase_Stopped.String() {
		client, err := r.cp.connect(node.status.GrpcEndpoint)
		if err != nil {
			return fmt.Errorf("connect to node %s failed %w", node.status.GrpcEndpoint, err)
		}

		err = client.stop()
		if err != nil {
			return fmt.Errorf("stop node %s failed %w", node.status.GrpcEndpoint, err)
		}
	}
	return nil
}

func isReadyNode(node node) bool {
	return node.status.Phase == report.Phase_Ready.String()
}

func isEnableNode(node node) bool {
	return node.info.Enable && node.status.Phase == report.Phase_Stopped.String()
}

func (r *nodeController) startNode(node node) error {
	client, err := r.cp.connect(node.status.GrpcEndpoint)
	if err != nil {
		return fmt.Errorf("connect to node %s failed %w", node.status.GrpcEndpoint, err)
	}

	err = client.start(node.info.Config)
	if err != nil {
		return fmt.Errorf("start node %s failed %w", node.status.GrpcEndpoint, err)
	}

	return nil
}

func needReloadNode(node node) bool {
	if node.status.Phase != report.Phase_Running.String() {
		return false
	}

	if node.status.Config != node.info.Config {
		return true
	}

	return false
}

func (r *nodeController) reloadNode(node node) error {
	client, err := r.cp.connect(node.status.GrpcEndpoint)
	if err != nil {
		return fmt.Errorf("connect to node %s failed %w", node.status.GrpcEndpoint, err)
	}

	err = client.reload(node.info.Config)
	if err != nil {
		return fmt.Errorf("start node %s failed %w", node.status.GrpcEndpoint, err)
	}

	return nil
}

func (n *nodeController) KeyProcessor(key etcd.Key) etcd.Key {
	return etcd.Base(key)
}

func (n *nodeController) ListenPrefix() []string {
	return []string{nodes.EtcdDirInfo, nodes.EtcdDirStatus}
}

type metrics struct {
	// total number of syncs
	SyncTotal metric.Int64Counter
	// error number of syncs
	SyncError metric.Int64Counter
}

func newMetrics() metrics {
	meter := otlp.Meter()

	syncTotal, _ := meter.Int64Counter("sync_total", metric.WithDescription("total number of syncs"))
	syncError, _ := meter.Int64Counter("sync_error", metric.WithDescription("error number of syncs"))

	return metrics{
		SyncTotal: syncTotal,
		SyncError: syncError,
	}
}
