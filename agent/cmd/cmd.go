package cmd

import (
	"xdp-banner/agent/cmd/global"
	"xdp-banner/agent/cmd/join"
	"xdp-banner/agent/cmd/server"

	"github.com/spf13/cobra"
)

func NewAgentCmd() *cobra.Command {
	opt := global.DefaultOption()

	cmd := &cobra.Command{
		Use:   "agent",
		Short: "xdp-banner Agent",
		Long:  `Agent is the daemon running on each node of the xdp-banner cluster to execute real work.`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	opt.SetFlags(cmd)
	cmd.AddCommand(join.Cmd(opt))
	cmd.AddCommand(server.Cmd(opt))

	return cmd
}
