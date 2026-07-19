package cmds

import (
	"encoding/csv"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func NewProjectAdd(app *App) *cobra.Command {
	return &cobra.Command{
		Use: "add",
		Run: func(cmd *cobra.Command, args []string) {
			projectPath, err := GetProjectPath(false)
			if err != nil {
				log.Fatal("cant get the project path")
			}

			projectCsv := filepath.Join(projectPath, "projects.csv")

			projectName := args[0]

			f, err := os.OpenFile(projectCsv, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0o600)
			if err != nil {
				log.Fatal("cant open project csv")
			}

			csvWriter := csv.NewWriter(f)
			csvWriter.Write([]string{projectName})
			csvWriter.Flush()
			os.Exit(0)
		},
	}
}
