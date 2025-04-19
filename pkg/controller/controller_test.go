package controller

import (
	"fmt"
	"sync"
	"testing"
	"time"
	"xdp-banner/pkg/etcd"
	"xdp-banner/pkg/informer"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type mockControllerImpl struct {
}

func (n *mockControllerImpl) SyncHandler(key string, informer informer.Informer) error {
	fmt.Println("SyncHandler", key)

	obj, exsit, err := informer.Get(etcd.Join("/agent/node/info", key))
	if err != nil {
		fmt.Println(err)
		return fmt.Errorf("get node info %s failed", key)
	} else if !exsit {
		fmt.Println("node info not exsit")
		return nil
	}

	infoStr, ok := obj.(string)
	if !ok {
		return fmt.Errorf("node info %s is not string", key)
	}

	fmt.Println("node info", infoStr)
	return fmt.Errorf("mock error")
}

func (n *mockControllerImpl) KeyProcessor(key etcd.Key) etcd.Key {
	return etcd.Base(key)
}

func (n *mockControllerImpl) ListenPrefix() []string {
	return []string{"/agent/node/info", "/agent/node/status"}
}

func TestController(t *testing.T) {
	etcdConfig := clientv3.Config{
		Endpoints:   []string{"etcd.joshua.su:2379"},
		DialTimeout: 5 * time.Second,
		Username:    "root",
		Password:    "SNKFFzL5h9",
	}

	cli, err := etcd.New(etcdConfig)
	if err != nil {
		t.Fatal(err)
	}

	deltaFIFO := informer.NewDeltaFIFOWithWait(2)

	reflectors := []*informer.Reflector{
		informer.NewReflector(cli, "agentnode_reflector", "/agent/node", deltaFIFO),
		informer.NewReflector(cli, "applied_reflector", "/agent/applied", deltaFIFO),
	}
	i := informer.New(deltaFIFO)

	go func() {
		wg := sync.WaitGroup{}

		for _, r := range reflectors {
			wg.Add(1)
			go r.Run(t.Context())
		}

		wg.Add(1)
		go i.Run(t.Context())

		wg.Wait()
	}()

	controller := New(i, ControllerOption{
		Name: "mock_controller",
		Impl: &mockControllerImpl{},
	})

	controller.Run(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	<-t.Context().Done()
}
