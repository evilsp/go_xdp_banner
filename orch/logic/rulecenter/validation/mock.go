package validation

import (
	"context"
	"xdp-banner/orch/model/rule"
)

type MockValidator struct {
}

func (v *MockValidator) Validate(ctx context.Context, c *rule.Rule) error {
	return nil
}
