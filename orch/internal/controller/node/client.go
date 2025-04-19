package node

import (
	"context"
	"fmt"
	"time"
	"xdp-banner/api/agent/v1/control"
	"xdp-banner/orch/server"

	"google.golang.org/grpc"
)

type clientPool map[string]*client

func newClientPool() clientPool {
	return make(clientPool)
}

func (c clientPool) connect(endpoint string) (client *client, err error) {
	client, ok := c[endpoint]
	if !ok {
		client, err = newClient(endpoint)
		if err != nil {
			return nil, err
		}
		c[endpoint] = client
	}

	return client, nil
}

type client struct {
	control control.ControlServiceClient
}

func newClient(endpoint string) (*client, error) {
	creds, err := server.NewCreditsInsecure()
	if err != nil {
		return nil, fmt.Errorf("failed to create credentials: %w", err)
	}

	c, err := grpc.NewClient(endpoint, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, err
	}
	control := control.NewControlServiceClient(c)

	return &client{
		control: control,
	}, nil
}

func runWithTimeout(op func(ctx context.Context) error, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return op(ctx)
}

func (c *client) start(configName string) error {
	op := func(ctx context.Context) error {
		_, err := c.control.Start(ctx, &control.StartRequest{
			ConfigName: configName,
		})
		return err
	}

	return runWithTimeout(op, 10*time.Second)
}

func (c *client) stop() error {
	op := func(ctx context.Context) error {
		_, err := c.control.Stop(ctx, nil)
		return err
	}

	return runWithTimeout(op, 10*time.Second)
}

func (c *client) reload(configName string) error {
	op := func(ctx context.Context) error {
		_, err := c.control.Reload(ctx, &control.ReloadRequest{
			ConfigName: configName,
		})
		return err
	}

	return runWithTimeout(op, 10*time.Second)
}
