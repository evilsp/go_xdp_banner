package global

import (
	"fmt"
	"xdp-banner/agent/internal/icert"

	"github.com/spf13/cobra"
)

type Option struct {
	Orch *OrchOption
}

func DefaultOption() *Option {
	return &Option{
		&OrchOption{
			Endpoints:   "127.0.0.1:6061",
			CAPath:      icert.CAFile,
			CertPath:    icert.CertFile,
			CertKeyPath: icert.KeyFile,
		},
	}
}

func (e *Option) Check() error {
	if err := e.Orch.Check(); err != nil {
		return err
	}

	return nil
}

func (e *Option) SetFlags(cmd *cobra.Command) {
	e.Orch.SetFlags(cmd)
}

type OrchOption struct {
	Endpoints   string
	CAPath      string
	CertPath    string
	CertKeyPath string
}

func (o *OrchOption) Check() error {
	if len(o.Endpoints) == 0 {
		return fmt.Errorf("xdp-banner endpoints is required")
	}

	return nil
}

func (o *OrchOption) SetFlags(cmd *cobra.Command) {
	cmdPrefix := "xdp-banner-"
	cmd.Flags().StringVarP(&o.Endpoints, cmdPrefix+"endpoints", "e", o.Endpoints, "xdp-banner orchestrator endpoints")
	cmd.Flags().StringVar(&o.CAPath, cmdPrefix+"capath", o.CAPath, "xdp-banner cluster ca path")
	cmd.Flags().StringVar(&o.CertPath, cmdPrefix+"certpath", o.CertPath, "local node cert path")
	cmd.Flags().StringVar(&o.CertKeyPath, cmdPrefix+"certkeypath", o.CertKeyPath, "local node cert key path")
}
