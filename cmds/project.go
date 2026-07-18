package cmds

import (
	"encoding/csv"
	"log"
	"os"
	"path/filepath"

	"github.com/fmo/pomodoro/internal/ui"
	"github.com/spf13/cobra"
)

func NewProjectCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "project",
		Short: "Project management",
		Run: func(cmd *cobra.Command, args []string) {
			projectPath, err := GetProjectPath(false)
			if err != nil {
				log.Fatal("cant get the project path")
			}

			projectCsv := filepath.Join(projectPath, "projects.csv")

			projectName := cmd.Flag("add").Value.String()

			if projectName != "" {
				f, err := os.OpenFile(projectCsv, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0o600)
				if err != nil {
					log.Fatal("cant open project csv")
				}

				csvWriter := csv.NewWriter(f)
				csvWriter.Write([]string{projectName})
				csvWriter.Flush()
				os.Exit(0)
			}

			if cmd.Flag("list") != nil {
				f, err := os.OpenFile(projectCsv, os.O_RDONLY, 0o600)
				if err != nil {
					log.Fatal("cant open project csv")
				}
				csvReader := csv.NewReader(f)
				records, err := csvReader.ReadAll()
				if err != nil {
					log.Fatal("cant read records")
				}
				ui.Render(records)
			}
		},
	}
}
