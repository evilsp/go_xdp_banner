package init

import (
	"xdp-banner/orch/cmd/global"

	"github.com/spf13/cobra"
)

type Option struct {
	Parent *global.ControllerOptions
}

func DefaultOption(parent *global.ControllerOptions) *Option {
	return &Option{
		Parent: parent,
	}
}

func (i *Option) Check() error {
	if err := i.Parent.Check(); err != nil {
		return err
	}

	return nil
}

func (i *Option) SetFlags(cmd *cobra.Command) {
}
