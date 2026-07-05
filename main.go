package main

import (
	"log"
	"log/slog"
	"os"

	"github.com/fmo/pomodoro/cmds"
	"github.com/spf13/viper"
)

func main() {
	// Project Path to add config path
	projectPath, err := cmds.GetProjectPath(true)
	if err != nil {
		log.Fatal("project path fetching failed")
	}

	// Logger File
	loggerFile, err := cmds.OpenFile("logger.log")
	if err != nil {
		log.Fatal(err)
	}

	// Logger
	handler := slog.NewTextHandler(loggerFile, &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
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
		viper.Set("pomodoro-file", "pomodoro.csv")
		viper.Set("backup-file", "pomodoro_bup.csv")
		viper.Set("backups", true)

		err = viper.WriteConfig()
		if err != nil {
			logger.Error("cant write config", "err", err)
		}
	}

	logger.Debug("should be read all the config", "pomodoro-file", viper.GetString("pomodoro-file"))

	// Pomodoro File
	openPomodoroFile, err := cmds.OpenFile(viper.GetString("pomodoro-file"))
	if err != nil {
		logger.Error("cant get the pomodoro file", "err", err)
		os.Exit(1)
	}

	pomodoroManager := cmds.NewPomodoroManager(openPomodoroFile)

	app := cmds.NewApp(logger, viper.GetViper(), pomodoroManager)

	cmds.Execute(app)
}
