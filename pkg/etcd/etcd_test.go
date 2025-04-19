package etcd

import (
	"context"
	"testing"
	"time"

	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	etcdConfig clientv3.Config = clientv3.Config{
		Endpoints:   []string{"10.43.4.33:2379"},
		DialTimeout: 5 * time.Second,
		Username:    "root",
		Password:    "0FJomdKTgi",
	}
)

func TestDelete(t *testing.T) {
	cli, err := New(etcdConfig)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := cli.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	ctx := context.Background()

	resp, err := cli.Delete(ctx, "not_exist")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(resp)
}

func TestList(t *testing.T) {
	cli, err := New(etcdConfig)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := cli.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	ctx := context.Background()
	opts := ListOption{
		Prefix: "/test",
		Size:   10,
	}

	r, err := cli.List(ctx, opts)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("total count: %d, total pages: %d, current page: %d, next cursor: %s", r.TotalCount, r.TotalPage, r.CurrentPage, r.NextCursor)
	for k, v := range r.Items.Iterator() {
		t.Logf("%s: %s", k, v)
	}

	opts.Cursor = r.NextCursor
	r, err = cli.List(ctx, opts)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("total count: %d, total pages: %d, current page: %d, next cursor: %s", r.TotalCount, r.TotalPage, r.CurrentPage, r.NextCursor)
	for k, v := range r.Items.Iterator() {
		t.Logf("%s: %s", k, v)
	}
}

func TestPager(t *testing.T) {
	cli, err := New(etcdConfig)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := cli.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	ctx := context.Background()
	opts := ListOption{
		Prefix: "/test",
		Size:   10,
	}

	pager := NewListPager(cli.List)
	pl, err := pager.List(ctx, opts)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("total count: %d, total pages: %d, current page: %d, next cursor: %s", pl.TotalCount, pl.TotalPage, pl.CurrentPage, pl.NextCursor)
	println(&pl.Items.Keys)
	println(pl.Items.Len())
	for k, v := range pl.Items.Iterator() {
		t.Logf("%s: %s", k, v)
	}
}

func TestKeepAliveOnce(t *testing.T) {
	cli, err := New(etcdConfig)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := cli.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	ctx := context.Background()

	resp, err := cli.Grant(ctx, 10)
	if err != nil {
		t.Fatal(err)
	}

	rwResp, err := cli.KeepAliveOnce(ctx, resp.ID)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(rwResp)

	rwResp1, err := cli.KeepAliveOnce(ctx, 114514)
	if err != nil {
		if err == rpctypes.ErrLeaseNotFound {
			t.Log(err)
		} else {
			t.Fatal(err)
		}
	}

	t.Log(rwResp1)
}

func TestWriteThorughCache(t *testing.T) {
	cli, err := New(etcdConfig)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := cli.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	ctx := context.Background()

	cache := NewWriteThroughCache(cli, 2, 10*time.Second, 2*time.Second)

	t.Log("First update, should update to etcd")
	err = cache.Update(ctx, "/test/status", "test_value")
	if err != nil {
		t.Fatal(err)
	}

	t.Log("Start to sleep for 5 seconds")
	time.Sleep(5 * time.Second)

	t.Log("Second update, should not update to etcd")
	err = cache.Update(ctx, "/test/status", "test_value")
	if err != nil {
		t.Fatal(err)
	}

	// 强制刷新
	t.Log("Third update, Force refresh")
	err = cache.Update(ctx, "/test/status", "test_value")
	if err != nil {
		t.Fatal(err)
	}

	t.Log("Fifth update, should not update to etcd")
	err = cache.Update(ctx, "/test/status", "test_value")
	if err != nil {
		t.Fatal(err)
	}

	t.Log("Sixth update, value changed ,should update to etcd")
	err = cache.Update(ctx, "/test/status", "test_value1")
	if err != nil {
		t.Fatal(err)
	}

	t.Log("Sleep for 15 seconds")
	time.Sleep(15 * time.Second)
	t.Log("Seventh update, cache miss ,should update to etcd")
	err = cache.Update(ctx, "/test/status", "test_value1")
	if err != nil {
		t.Fatal(err)
	}

}

func TestRevoke(t *testing.T) {
	cli, err := New(etcdConfig)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := cli.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	ctx := context.Background()

	resp, err := cli.Grant(ctx, 10)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(resp.ID)
	rwResp, err := cli.Revoke(ctx, resp.ID)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(rwResp)

	rwResp1, err := cli.Revoke(ctx, 114514)
	if err != nil {
		if err == rpctypes.ErrLeaseNotFound {
			t.Log(err)
			t.Log(rwResp1)
		} else {
			t.Fatal(err)
		}
	}

	t.Log(rwResp1)
}

func TestWatch(t *testing.T) {
	etcdConfig.Endpoints = []string{"114514"}
	cli, err := New(etcdConfig)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := cli.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	watchController, err := cli.Watch(WatchOption{
		Prefix: "/test",
	})
	defer watchController.Stop()

	if err != nil {
		t.Fatal(err)
	}

	for event := range watchController.ResultChan() {
		switch event.Type {
		case Put:
			t.Logf("key: %s, value: %s", event.Key, event.Value)
		case Delete:
			t.Logf("key: %s", event.Key)
		case Error:
			t.Logf("error: %s", event.Value)
		default:
			t.Log("unknown event type")
		}
	}
}
