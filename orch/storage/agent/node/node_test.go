package node

import (
	"context"
	"sync"
	"testing"
	"time"
	"xdp-banner/orch/model/node"
	"xdp-banner/pkg/etcd"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func TestStatus(t *testing.T) {
	client, err := etcd.New(
		clientv3.Config{
			Endpoints: []string{"etcd.joshua.su:2379"},
			Username:  "root",
			Password:  "SNKFFzL5h9",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	status := NewStatusStorage(context.Background(), client)

	wg := sync.WaitGroup{}
	errChan := make(chan error, 3)

	// listen status change
	wg.Add(1)
	go func() {
		defer wg.Done()

		event, err := client.Watch(etcd.WatchOption{
			Prefix: EtcdDirStatus,
		})
		if err != nil {
			errChan <- err
			return
		}

		for {
			select {
			case <-t.Context().Done():
				event.Stop()
				return
			case e := <-event.ResultChan():
				t.Logf("event: %v", e)
			}
		}
	}()

	updateStatus := func(ctx context.Context, statusObj *node.AgentStatus) error {
		return loopIntervalWithTimes(ctx, 5*time.Second, 3, func() error {
			return status.Update(t.Context(), statusObj.Name, statusObj)
		})
	}

	wg.Add(1)
	go func() {
		defer wg.Done()

		statusObj := &node.AgentStatus{
			CommonStatus: node.CommonStatus{
				Name: "test",
			},
			GrpcEndpoint: "example.com:6063",
			Config:       "default",
			Phase:        "Ready",
		}

		err := updateStatus(t.Context(), statusObj)
		if err != nil {
			errChan <- err
			return
		}

		statusObj.Phase = "Running"
		err = updateStatus(t.Context(), statusObj)
		if err != nil {
			errChan <- err
			return
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		statusObj := &node.AgentStatus{
			CommonStatus: node.CommonStatus{
				Name: "test1",
			},
			GrpcEndpoint: "example.com:6063",
			Config:       "default",
			Phase:        "Ready",
		}

		err := updateStatus(t.Context(), statusObj)
		if err != nil {
			errChan <- err
			return
		}

		statusObj.Phase = "Running"
		err = updateStatus(t.Context(), statusObj)
		if err != nil {
			errChan <- err
			return
		}
	}()

	getStatus := func(ctx context.Context, name string) error {
		statusObj, err := status.Get(ctx, name, false)
		if err != nil {
			if err == ErrStatusNotFound {
				t.Logf("status %s not found", name)
				return nil
			}
			return err
		}

		t.Logf("status: %v", statusObj)
		return nil
	}

	wg.Add(1)
	go func() {
		defer wg.Done()

		err := loopIntervalWithTimes(t.Context(), 2*time.Second, 15, func() error {
			if err := getStatus(t.Context(), "test"); err != nil {
				return err
			}
			if err := getStatus(t.Context(), "test1"); err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			errChan <- err
			return
		}
	}()

	select {
	case err := <-errChan:
		t.Fatal(err)
	case <-t.Context().Done():
	}
	wg.Wait()
}

func loopIntervalWithTimes(ctx context.Context, interval time.Duration, times int, f func() error) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range times {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			err := f()
			if err != nil {
				return err
			}
		}
	}

	return nil
}
