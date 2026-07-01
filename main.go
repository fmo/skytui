package main

import (
	"log/slog"
	"os"
	"path/filepath"

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

	csvFile, err := cmd.GetCsvFile()
	if err != nil {
		logger.Error("cant get the csv file", "err", err)
		os.Exit(1)
	}

	viper.WithLogger(logger)
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath(projectPath)

	configFile := filepath.Join(projectPath, "config.yml")

	_, err = os.Open(configFile)
	if err != nil {
		_, err = os.Create(configFile)
		if err != nil {
			logger.Error("cant create config", "err", err)
		}
	}
	if os.Getenv("env") != "" {
		viper.Set("env", "dev")
	}
	viper.Set("csv", "pomodoro.csv")

	err = viper.WriteConfig()
	if err != nil {
		logger.Error("cant write config", "err", err)
	}

	if err := viper.ReadInConfig(); err != nil {
		logger.Error("cant read config file", "err", err)
		os.Exit(1)
	}

	app := cmd.NewApp(logger, viper.GetViper(), csvFile)

	cmd.Execute(app)
}
