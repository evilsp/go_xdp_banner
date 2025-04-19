package informer

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
	"xdp-banner/pkg/etcd"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type MockHandler struct {
}

func (h *MockHandler) OnAdd(key string, obj any, isInInitialList bool) {
	fmt.Println("OnAdd", key, obj, isInInitialList)

}
func (h *MockHandler) OnUpdate(key string, oldObj, newObj any) {
	fmt.Println("OnUpdate", key, oldObj, newObj)

}
func (h *MockHandler) OnDelete(key string, obj any) {
	fmt.Println("OnDelete", key, obj)
}

func TestInformer(t *testing.T) {
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

	deltaFIFO := NewDeltaFIFOWithWait(2)

	reflectors := []*Reflector{
		NewReflector(cli, "agentnode_reflector", "/agent/node", deltaFIFO),
		NewReflector(cli, "applied_reflector", "/agent/applied", deltaFIFO),
	}
	i := New(deltaFIFO)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	m := &MockHandler{}
	i.Register([]string{"/agent/node/info", "/agent/node/register"}, m)

	wg := sync.WaitGroup{}

	for _, r := range reflectors {
		wg.Add(1)
		go r.Run(ctx)
	}

	wg.Add(1)
	go i.Run(ctx)

	wg.Wait()
}
