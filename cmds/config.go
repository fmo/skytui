package cmds

import "github.com/spf13/cobra"

func NewConfigCmd() *cobra.Command {
	return &cobra.Command{
		Use: "config",
	}
}
