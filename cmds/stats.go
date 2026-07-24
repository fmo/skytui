package cmds

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

func NewStatsCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "stats",
		Short: "Pomodoro stats",
		Run: func(cmd *cobra.Command, args []string) {
			projectPath, err := GetProjectPath()
			if err != nil {
				app.logger.Error("cant get the project path", "err", err)
				os.Exit(1)
			}

			fullPath := filepath.Join(projectPath, "pomodoro.csv")

			f, err := os.Open(fullPath)
			if err != nil {
				app.logger.Error("cant open the pomodoro.csv file", "err", err)
				os.Exit(1)
			}

			csvReader := csv.NewReader(f)
			records, err := csvReader.ReadAll()
			if err != nil {
				app.logger.Error("cant read the pomodoro records", "err", err)
				os.Exit(1)
			}

			var total time.Duration
			for _, record := range records {
				d, err := time.ParseDuration(record[1])
				if err != nil {
					app.logger.Error("cant parse the duration", "err", err)
					os.Exit(1)
				}
				total += d
			}

			fmt.Printf("total duration: %s", total.String())
			os.Exit(0)
		},
	}
}
