package cmds

import (
	"encoding/csv"
	"log"
	"os"
	"path/filepath"

	"github.com/fmo/skytui/internal/ui"
	"github.com/spf13/cobra"
)

func NewProjectList(app *App) *cobra.Command {
	return &cobra.Command{
		Use: "list",
		Run: func(cmd *cobra.Command, args []string) {
			projectPath, err := GetProjectPath()
			if err != nil {
				log.Fatal("cant get the project path")
			}

			projectCsv := filepath.Join(projectPath, "projects.csv")

			f, err := os.OpenFile(projectCsv, os.O_RDONLY, 0o600)
			if err != nil {
				log.Fatal("cant open project csv")
			}
			csvReader := csv.NewReader(f)
			records, err := csvReader.ReadAll()
			if err != nil {
				log.Fatal("cant read records")
			}
			if err := ui.Render(records); err != nil {
				app.logger.Error("cant render records", "err", err)
			}
		},
	}
}
