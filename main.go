package main

import (
	"log/slog"
	"os"

	"github.com/fmo/pomodoro/cmd"
	"github.com/spf13/viper"
)

func main() {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	})
	logger := slog.New(handler)

	projectPath, err := cmd.GetProjectPath(true)
	if err != nil {
		logger.Error("project path fetching failed", "err", err)
		os.Exit(1)
	}

	openPomodoroFile, err := cmd.OpenPomodoroFile()
	if err != nil {
		logger.Error("cant get the pomodoro file", "err", err)
		os.Exit(1)
	}

	err = cmd.OpenConfigFile()
	if err != nil {
		logger.Error("cant open config file", "err", err)
	}

	viper.WithLogger(logger)
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath(projectPath)

	if os.Getenv("env") != "" {
		viper.Set("env", "dev")
	}
	viper.Set("csv", "pomodoro.csv")
	viper.Set("backups", true)

	err = viper.WriteConfig()
	if err != nil {
		logger.Error("cant write config", "err", err)
	}

	if err := viper.ReadInConfig(); err != nil {
		logger.Error("cant read config file", "err", err)
		os.Exit(1)
	}

	app := cmd.NewApp(logger, viper.GetViper(), openPomodoroFile)

	cmd.Execute(app)
}
