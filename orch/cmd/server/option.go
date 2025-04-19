package server

import (
	"fmt"
	"xdp-banner/orch/cmd/global"
	"xdp-banner/pkg/option"

	"github.com/spf13/cobra"
)

type Option struct {
	Parent *global.ControllerOptions

	GrpcAddr string
	HttpAddr string
	Otlp     *option.OtlpOption
}

func DefaultOption(parent *global.ControllerOptions) *Option {
	return &Option{
		Parent:   parent,
		GrpcAddr: "0.0.0.0:6061",
		HttpAddr: "0.0.0.0:6062",
		Otlp:     option.DefaultOtelOption(),
	}
}

func (o *Option) Check() error {
	if err := o.Parent.Check(); err != nil {
		return err
	}

	if o.GrpcAddr == "" {
		return fmt.Errorf("grpc addr is empty")
	}

	return nil
}

func (o *Option) SetFlags(cmd *cobra.Command) {
	o.Parent.SetFlags(cmd)
	o.Otlp.SetFlags(cmd)

	cmd.Flags().StringVar(&o.GrpcAddr, "server.grpc-addr", o.GrpcAddr, "set grpc address")
	cmd.Flags().StringVar(&o.HttpAddr, "server.http-addr", o.HttpAddr, "set http address")
}
