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
	// Create project folder if it does not exist
	if err := cmds.CreateProjectPath(); err != nil {
		log.Fatal("can't create project path")
	}

	// Project path to add config path
	projectPath, err := cmds.GetProjectPath()
	if err != nil {
		log.Fatal("project path fetching failed")
	}

	// Logger file
	loggerFile, err := os.OpenFile(filepath.Join(projectPath, cmds.LogFile), os.O_CREATE|os.O_RDWR|os.O_APPEND, 0o600)
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

	// Viper setup
	viper.AddConfigPath(projectPath)
	if err := viper.ReadInConfig(); err != nil {
		_, ok := err.(viper.ConfigFileNotFoundError)
		if ok {
			viper.Set("pomodoro-file", cmds.PomodoroFile)
			if err := viper.WriteConfigAs(filepath.Join(projectPath, cmds.ConfigFile)); err != nil {
				logger.Error("cant write the config to config file", "err", err)
				os.Exit(1)
			}
		} else {
			logger.Error("cant read config", "err", err)
			os.Exit(1)
		}
	}

	// Run the app
	app := cmds.NewApp(logger, viper.GetViper())
	cmds.Execute(app)
}
