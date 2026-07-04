package cmds

import (
	"log/slog"

	"github.com/spf13/viper"
)

type App struct {
	logger          *slog.Logger
	viper           *viper.Viper
	pomodoroManager *PomodoroManager
}

func NewApp(logger *slog.Logger, viper *viper.Viper, pomodoroManager *PomodoroManager) *App {
	return &App{logger, viper, pomodoroManager}
}

func (app *App) SavePomodoro(limit, count int) error {
	if app.viper != nil {
		bup := app.viper.GetBool("backups")

		if bup {
			app.logger.Debug("creating backups")
			app.pomodoroManager.CreatePomodoroBackup(app.viper.GetString("backup-file"))
		}
	}

	return app.pomodoroManager.Save(limit, count)
}
