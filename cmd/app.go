package cmd

import (
	"encoding/csv"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/spf13/viper"
)

type App struct {
	logger       *slog.Logger
	viper        *viper.Viper
	pomodoroFile *os.File
}

func NewApp(logger *slog.Logger, viper *viper.Viper, csvFile *os.File) *App {
	return &App{logger, viper, csvFile}
}

func (app *App) SavePomodoro(limit, count int) error {
	if app.viper != nil {
		if bup, ok := app.viper.Get("backups").(bool); ok {
			if bup {
				app.logger.Debug("creating backups")
				app.CreatePomodoroBackup()
			}
		}
	}

	if app.pomodoroFile == nil {
		app.logger.Error("pomodoro file is nil")
		os.Exit(1)
	}

	w := csv.NewWriter(app.pomodoroFile)

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

func (app *App) CreatePomodoroBackup() error {
	src, err := os.Open(app.pomodoroFile.Name())
	if err != nil {
		return err
	}

	fp, err := GetProjectPath(false)
	if err != nil {
		return err
	}
	csvFile := app.viper.GetString("backup-file")

	dsc, err := os.Create(filepath.Join(fp, csvFile))
	if err != nil {
		return err
	}

	io.Copy(dsc, src)

	return nil
}
