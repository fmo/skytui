package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/fmo/skytui/cmds"
	"github.com/spf13/viper"
)

func main() {
	// Project Path to add config path
	projectPath, err := cmds.GetProjectPath(true)
	if err != nil {
		log.Fatal("project path fetching failed")
	}

	loggerPath := filepath.Join(projectPath, "logger.log")

	// Logger File
	loggerFile, err := os.OpenFile(loggerPath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0o600)
	if err != nil {
		log.Fatalf("cant open logger")
	}

	// Logger
	handler := slog.NewTextHandler(loggerFile, &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == "source" {
				if source, ok := a.Value.Any().(*slog.Source); ok {
					sourceParts := strings.Split(source.File, "/")
					sourcePartsLen := len(sourceParts)
					a.Value = slog.StringValue(fmt.Sprintf("%s:%d", sourceParts[sourcePartsLen-1], source.Line))
				}
			}
			return a
		},
	})
	logger := slog.New(handler)

	// Create Config if does not exist
	_, err = cmds.OpenFile("config.yml")
	if err != nil {
		logger.Error("cant open config file", "err", err)
	}

	viper.WithLogger(logger)
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath(projectPath)

	if err := viper.ReadInConfig(); err != nil {
		logger.Error("cant read config file", "err", err)
		os.Exit(1)
	}

	if !viper.InConfig("pomodoro-file") {
		viper.Set("pomodoro-file", "pomodoro.csv")
	}
	if !viper.InConfig("backup-file") {
		viper.Set("backup-file", "pomodoro_bup.csv")
	}
	if !viper.InConfig("backups") {
		viper.Set("backups", true)
	}

	err = viper.WriteConfig()
	if err != nil {
		logger.Error("cant write config", "err", err)
		os.Exit(1)
	}

	logger.Info("loaded all configuration")

	// Pomodoro File
	openPomodoroFile, err := cmds.OpenFile(viper.GetString("pomodoro-file"))
	if err != nil {
		logger.Error("cant get the pomodoro file", "err", err)
		os.Exit(1)
	}

	pomodoroManager := cmds.NewPomodoroManager(openPomodoroFile, logger)

	app := cmds.NewApp(logger, viper.GetViper(), pomodoroManager)

	cmds.Execute(app)
}
