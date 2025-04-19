package server

import (
	"fmt"
	"time"
	"xdp-banner/agent/cmd/global"
	"xdp-banner/pkg/option"

	"github.com/spf13/cobra"
)

type Option struct {
	Parent *global.Option

	GrpcAddr       string
	ReportInterval time.Duration
	Otlp           *option.OtlpOption
}

func DefaultOption(parent *global.Option) *Option {
	return &Option{
		Parent:         parent,
		GrpcAddr:       "0.0.0.0:6063",
		ReportInterval: 15 * time.Second,
		Otlp:           option.DefaultOtelOption(),
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

	cmd.Flags().StringVar(&o.GrpcAddr, "grpc-addr", o.GrpcAddr, "grpc server address")
	cmd.Flags().DurationVar(&o.ReportInterval, "report-interval", o.ReportInterval, "set agent report status interval to the orch")
}
