package main

import (
	"log/slog"
	"os"

	"github.com/fmo/pomodoro/cmd"
	"github.com/spf13/viper"
)

func main() {
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath(".")

	viper.ReadInConfig()

	handler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)

	app := cmd.NewApp(logger, viper.GetViper())

	cmd.Execute(app)
}
