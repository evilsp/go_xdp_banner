package validation

import (
	"context"
	"xdp-banner/orch/model/rule"
)

// Validator validates the configuration.
type Validator interface {
	// Validate validates the configuration.
	Validate(ctx context.Context, c *rule.Rule) error
}
