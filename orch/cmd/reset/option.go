package reset

import (
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

	return nil
}

func (o *Option) SetFlags(cmd *cobra.Command) {
}
