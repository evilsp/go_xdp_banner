package cmd

import (
	"fmt"
	"xdp-banner/orch/cmd/global"
	initCluster "xdp-banner/orch/cmd/init"
	"xdp-banner/orch/cmd/join"
	"xdp-banner/orch/cmd/reset"
	"xdp-banner/orch/cmd/server"

	"github.com/spf13/cobra"
)

var (
	cfgFile string // 用来接收 --config
)

func NewOrchCmd() *cobra.Command {
	opt := global.DefaultOption()

	cmd := &cobra.Command{
		Use:   "orch",
		Short: "xdp-banner Orchestrator",
		Long:  `Orch is the control plane for the xdp-banner cluster.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// 加载文件
			vpr, loadedOpts, err := global.LoadConfig(cfgFile)
			if err != nil {
				return fmt.Errorf("failed to load config %q: %w", cfgFile, err)
			}
			// 用用户文件里的配置 覆盖 掉 opt 的默认值
			*opt = *loadedOpts

			// （如果你需要 viper 实例，也可以把 vpr 赋给包级变量）
			_ = vpr
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	// 3) 在顶级命令上注册 --config / -c flag
	cmd.PersistentFlags().StringVarP(
		&cfgFile, "config", "c", "./config.yaml",
		"config file (default is ./config.yaml)",
	)

	opt.SetFlags(cmd)
	cmd.AddCommand(initCluster.Cmd(opt))
	cmd.AddCommand(join.Cmd(opt))
	cmd.AddCommand(reset.Cmd(opt))
	cmd.AddCommand(server.Cmd(opt))

	return cmd
}
