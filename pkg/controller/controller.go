package controller

import (
	"context"
	"fmt"
	"time"
	"xdp-banner/pkg/etcd"
	"xdp-banner/pkg/informer"
	"xdp-banner/pkg/log"
	"xdp-banner/pkg/queue"
	"xdp-banner/pkg/wait"
)

type Controller struct {
	name     string
	opt      ControllerOption
	informer informer.Informer
	impl     ControllerImpl
	queue    queue.RateLimitingQueue[etcd.Key]
	logger   *log.Logger
}

type ControllerOption struct {
	Name string
	Impl ControllerImpl
}

type ControllerImpl interface {
	SyncHandler(key string, informer informer.Informer) error
	KeyProcessor(key etcd.Key) etcd.Key
	ListenPrefix() []string
}

func New(informer informer.Informer, opt ControllerOption) *Controller {
	controller := &Controller{
		name:     opt.Name,
		opt:      opt,
		informer: informer,
		impl:     opt.Impl,
		queue:    queue.NewRateLimitingQueueWithName(queue.DefaultControllerRateLimiter[string](), opt.Name),
		logger:   log.With(log.StringField("controller", opt.Name)),
	}

	if controller.impl == nil {
		controller.impl = &noOpControllerImpl{}
	}

	informer.Register(controller.impl.ListenPrefix(), controller)
	return controller
}

func (c *Controller) Run(ctx context.Context) {
	c.logger.Info("Starting the controller")
	c.logger.Info("Waiting for the informer caches to sync")

	c.informer.WaitForCacheSync(ctx)
	c.logger.Info("Informer caches are synced")

	go wait.Until(c.runWorker, time.Second, ctx)
	c.logger.Info("Started workers")

	<-ctx.Done()
	c.stop()
}

func (c *Controller) stop() {
	c.queue.ShutDown()
	c.logger.Info("controller stopped")
}

func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

// 取数据处理
func (c *Controller) processNextWorkItem() bool {
	name, shutdown := c.queue.Get()
	if shutdown {
		return false
	}

	err := func(name etcd.Key) error {
		defer c.queue.Done(name)

		if err := c.impl.SyncHandler(name, c.informer); err != nil {
			c.queue.AddRateLimited(name)
			return fmt.Errorf("Controller error syncing '%s': %s, requeuing", name, err.Error())
		}

		c.queue.Forget(name)
		c.logger.Info("Controller successfully synced", log.StringField("key", name))
		return nil
	}(name)

	if err != nil {
		c.logger.Error("error syncing", log.StringField("name", name), log.ErrorField(err))
		return true
	}
	return true
}

func (c *Controller) OnAdd(key etcd.Key, obj any, isInInitialList bool) {
	c.enqueue(key)
}

func (c *Controller) OnUpdate(key etcd.Key, oldObj, newObj any) {
	c.enqueue(key)
}

func (c *Controller) OnDelete(key etcd.Key, obj any) {
	c.enqueue(key)
}

func (c *Controller) enqueue(key etcd.Key) {
	key = c.impl.KeyProcessor(key)
	if key == "" {
		c.logger.Warn("obj added, but key is empty", log.StringField("key", key))
		return
	}

	c.queue.Add(key)
}

// noOpControllerImpl is a no-op implementation of ControllerImpl
type noOpControllerImpl struct {
}

func (n *noOpControllerImpl) SyncHandler(key string, informer informer.Informer) error {
	return nil
}

func (n *noOpControllerImpl) KeyProcessor(key etcd.Key) etcd.Key {
	return key
}

func (n *noOpControllerImpl) ListenPrefix() []string {
	return nil
}
