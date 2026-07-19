package cmds

import (
	"github.com/spf13/cobra"
)

func NewProjectCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Project management",
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	projectAddCmd := NewProjectAdd(app)
	projectListCmd := NewProjectList(app)
	projectDefaultCmd := NewProjectDefault(app)
	cmd.AddCommand(projectAddCmd)
	cmd.AddCommand(projectListCmd)
	cmd.AddCommand(projectDefaultCmd)

	return cmd
}
