package reset

import (
	"context"
	"fmt"
	"time"
	"xdp-banner/orch/cmd/global"
	"xdp-banner/orch/storage"
	"xdp-banner/pkg/log"

	icert "xdp-banner/orch/internal/cert"

	"github.com/spf13/cobra"

	"go.uber.org/zap"
)

func Cmd(parentOpt *global.ControllerOptions) *cobra.Command {
	opt := DefaultOption(parentOpt)

	cmd := &cobra.Command{
		Use:   "reset",
		Short: "reset the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !promtCheck() {
				return nil
			}

			if err := opt.Check(); err != nil {
				return err
			}

			resetCluster(opt)
			return nil
		},
	}

	opt.SetFlags(cmd)

	return cmd
}

func promtCheck() bool {
	fmt.Println("This operation will reset the cluster, all data will be lost, continue? [y/N]")
	var input string
	if _, err := fmt.Scanln(&input); err != nil {
		log.Fatal("Read input failed", zap.String("err", err.Error()))
	}
	if input != "y" && input != "Y" {
		return false
	}

	return true
}

func resetCluster(opt *Option) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Cli init
	err := global.CreateGlobalEtcdInstance(opt.Parent)

	if err != nil {
		log.Fatal("Connect etcd failed", zap.String("err", err.Error()))
	}

	s := storage.New(ctx, global.Cli)

	log.Info("Resetting cluster")

	errs := []error{}

	/*
	 Remote
	*/
	log.Info("Deleting all agent")
	if err := s.Agent.DeleteDir(ctx); err != nil {
		errs = append(errs, fmt.Errorf("failed to delete nodes dir: %w", err))
	}

	log.Info("Deleting all orch")
	if err := s.Orch.DeleteDir(ctx); err != nil {
		errs = append(errs, fmt.Errorf("failed to delete cluster cert: %w", err))
	}

	/*
	 Local
	*/
	if err := icert.DeleteLocalCertDir(); err != nil {
		errs = append(errs, fmt.Errorf("failed to delete local cert dir: %w", err))
	}

	if len(errs) == 0 {
		log.Info("Cluster reset successfully")
	} else {
		for _, err := range errs {
			log.Error("Cluster reset failed", zap.String("err", err.Error()))
		}
	}
}
