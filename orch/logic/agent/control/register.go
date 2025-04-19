package control

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	model "xdp-banner/orch/model/node"
	"xdp-banner/orch/storage/agent/node"
	"xdp-banner/pkg/errors"
)

func (c *Control) RegisterNode(ctx context.Context, name string) (string, error) {
	token, err := generateToken()
	if err != nil {
		errors.NewServiceErrorf("failed to generate token: %v", err)
	}

	rm := &model.Registration{
		Name:  name,
		Token: token,
	}

	err = c.registers.Add(ctx, rm)
	if err != nil {
		if err == node.ErrRegisterExist {
			return "", errors.NewInputError("node already registered")
		}
		return "", errors.NewServiceErrorf("failed to register node: %v", err)
	}

	return token, nil
}

// generateToken generates a random token
func generateToken() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (c *Control) UnRegisterNode(ctx context.Context, name string) error {
	err := c.registers.Delete(ctx, name)
	if err != nil {
		return errors.NewServiceErrorf("failed to unregister node: %v", err)
	}

	return nil
}

func (c *Control) ListRegistration(ctx context.Context, size int64, nextCursor string) (model.RegistrationList, error) {
	registrations, err := c.registers.List(ctx, size, nextCursor)
	if err != nil {
		return model.RegistrationList{}, errors.NewServiceErrorf("failed to list registrations: %v", err)
	}

	return registrations, nil
}
