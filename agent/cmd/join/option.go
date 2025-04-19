package join

import (
	"errors"
	"time"
	"xdp-banner/agent/cmd/global"

	"github.com/spf13/cobra"
)

type Option struct {
	Token   string
	Timeout time.Duration

	Parent *global.Option
}

func DefaultOption(parent *global.Option) *Option {
	return &Option{
		Parent:  parent,
		Timeout: 30 * time.Second,
	}
}

func (o *Option) Check() error {
	if err := o.Parent.Check(); err != nil {
		return err
	}

	if o.Token == "" {
		return errors.New("token is required")
	}

	return nil
}

func (o *Option) SetFlags(cmd *cobra.Command) {
	o.Parent.SetFlags(cmd)

	cmd.Flags().StringVarP(&o.Token, "token", "t", o.Token, "token for join")
	cmd.Flags().DurationVarP(&o.Timeout, "timeout", "T", o.Timeout, "timeout for join")
}
