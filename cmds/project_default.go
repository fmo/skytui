package cmds

import (
	"encoding/csv"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func NewProjectDefault(app *App) *cobra.Command {
	return &cobra.Command{
		Use: "default",
		Run: func(cmd *cobra.Command, args []string) {
			defaultProject := args[0]
			projectPath, err := GetProjectPath()
			fullPath := filepath.Join(projectPath, "projects.csv")

			fileToRead, err := os.Open(fullPath)
			if err != nil {
				app.logger.Error("cant read project file", "err", err)
				os.Exit(1)
			}

			r := csv.NewReader(fileToRead)

			projects, err := r.ReadAll()
			if err != nil {
				app.logger.Error("cant read projects file", "err", err)
				os.Exit(1)
			}

			fileToWrite, err := os.OpenFile(fullPath, os.O_TRUNC|os.O_RDWR, 0o600)
			if err != nil {
				app.logger.Error("cant open project file", "err", err)
				os.Exit(1)
			}

			w := csv.NewWriter(fileToWrite)

			for _, project := range projects {
				if project[0] == defaultProject {
					w.Write([]string{project[0], "default"})
					continue
				}
				w.Write([]string{project[0], ""})
			}
			w.Flush()
		},
	}
}
