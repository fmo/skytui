package cmd

import (
	"encoding/csv"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/spf13/viper"
)

type App struct {
	logger  *slog.Logger
	viper   *viper.Viper
	csvFile *os.File
}

func NewApp(logger *slog.Logger, viper *viper.Viper, csvFile *os.File) *App {
	return &App{logger, viper, csvFile}
}

func (app *App) Save(limit, count int) error {
	if app.viper != nil {
		if bup, ok := app.viper.Get("app.backup-active").(bool); ok {
			if bup {
				CreateBackup(app.csvFile)
			}
		}
	}

	w := csv.NewWriter(app.csvFile)

	// completed is total duration in seconds - count down time, ie total 30 seconds - 10 count down (left)
	completed := time.Duration(limit-count) * time.Second

	left := strconv.Itoa(count)

	// current time, completed time, count down
	err := w.Write([]string{time.Now().Format(time.RFC3339), completed.String(), fmt.Sprintf("%ss", left)})
	if err != nil {
		return err
	}
	w.Flush()

	return nil
}
