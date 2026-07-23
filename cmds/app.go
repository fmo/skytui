package cmds

import (
	"log/slog"

	"github.com/spf13/viper"
)

type App struct {
	logger *slog.Logger
	viper  *viper.Viper
}

func NewApp(logger *slog.Logger, viper *viper.Viper) *App {
	return &App{logger, viper}
}
