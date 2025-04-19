package informer

import (
	"context"
	"testing"
	"time"
	"xdp-banner/pkg/etcd"

	"log"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type MockStore struct{}

func (m *MockStore) Add(key string, value interface{}) error {
	log.Printf("add key: %s, value: %v\n", key, value)
	return nil
}

func (m *MockStore) Update(key string, value interface{}) error {
	log.Printf("update key: %s, value: %v\n", key, value)
	return nil
}

func (m *MockStore) Delete(key string) error {
	log.Printf("delete key: %s\n", key)
	return nil
}

func (m *MockStore) Replace(iter ListIter, version string) error {
	log.Printf("replace, version: %s\n", version)
	for k, v := range iter {
		log.Printf("key: %s, value: %v\n", k, v)
	}
	return nil
}

func (m *MockStore) List() []interface{} {
	log.Println("not implemented")
	return nil
}

func (m *MockStore) ListKeys() []string {
	log.Println("not implemented")
	return nil
}

func (m *MockStore) Get(key string) (item interface{}, exists bool, err error) {
	log.Println("not implemented")
	return nil, false, nil
}

func (m *MockStore) SyncDone() {
	log.Println("sync done")
}

func TestReflector(t *testing.T) {

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

	r := NewReflector(cli, "node", "/test/", &MockStore{})

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(15 * time.Second)
		cancel()
	}()

	r.Run(ctx)
}
