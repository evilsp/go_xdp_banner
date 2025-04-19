package join

import (
	"errors"
	"xdp-banner/orch/cmd/global"

	"github.com/spf13/cobra"
)

type Option struct {
	Token string

	Parent *global.ControllerOptions
}

func DefaultOption(parent *global.ControllerOptions) *Option {
	return &Option{
		Parent: parent,
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
	cmd.Flags().StringVarP(&o.Token, "token", "t", o.Token, "cluster token")
}
