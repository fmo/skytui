package cmds

import (
	"encoding/csv"
	"log"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type App struct {
	logger          *slog.Logger
	viper           *viper.Viper
	pomodoroManager *PomodoroManager
	defaultProject  string
}

func NewApp(logger *slog.Logger, viper *viper.Viper, pomodoroManager *PomodoroManager) *App {
	projectPath, err := GetProjectPath(false)
	if err != nil {
		log.Fatal("cant get the project path")
	}
	fullPath := filepath.Join(projectPath, "projects.csv")

	f, err := os.Open(fullPath)
	if err != nil {
		log.Fatal("cant open the path")
	}

	r := csv.NewReader(f)
	projects, err := r.ReadAll()
	if err != nil {
		log.Fatal("cant get the projects")
	}

	defaultProject := ""
	for _, project := range projects {
		if len(project) == 2 && project[1] == "default" {
			defaultProject = project[0]
		}
	}

	return &App{logger, viper, pomodoroManager, defaultProject}
}

func (app *App) SavePomodoro(limit, count int) error {
	if app.viper != nil {
		bup := app.viper.GetBool("backups")

		if bup {
			app.logger.Info("creating backup from the previous pomodoro records")
			app.pomodoroManager.CreatePomodoroBackup(app.viper.GetString("backup-file"))
		}
	}

	return app.pomodoroManager.Save(limit, count)
}
