package cmd

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/cobra"
)

func NewStatsCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "stats",
		Short: "Your full focus time",
		Long:  "See how much foucsed time you had during the day",
		Run: func(cmd *cobra.Command, args []string) {
			period, err := cmd.Flags().GetString("period")
			if err != nil {
				app.logger.Error("need period to show stats", "err", err)
				os.Exit(1)
			}

			file, err := os.Open(app.pomodoroFile.Name())
			if err != nil {
				app.logger.Error("cant open csv file", "err", err)
				os.Exit(1)
			}

			reader := csv.NewReader(file)
			records, err := reader.ReadAll()
			if err != nil {
				app.logger.Error("cant read the records", "err", err)
				os.Exit(1)
			}

			var total time.Duration
			switch period {
			case "today":
				for _, r := range records {
					d, err := time.ParseDuration(r[1])
					if err != nil {
						log.Fatal(err)
					}
					recordDate, err := time.Parse(time.RFC3339, r[0])
					if err != nil {
						log.Fatal(err)
					}
					if time.Now().Day() == recordDate.Day() && time.Now().Month() == recordDate.Month() && time.Now().Year() == recordDate.Year() {
						total += d
					}
				}
			case "yesterday":
				for _, r := range records {
					d, err := time.ParseDuration(r[1])
					if err != nil {
						log.Fatal(err)
					}
					recordDate, err := time.Parse(time.RFC3339, r[0])
					if err != nil {
						log.Fatal(err)
					}
					if time.Now().Day()-1 == recordDate.Day() && time.Now().Month() == recordDate.Month() && time.Now().Year() == recordDate.Year() {
						total += d
					}
				}

			}

			fmt.Printf("%s\n", total.String())
		},
	}
}
