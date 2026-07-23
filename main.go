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
	// Create Project Path
	if err := cmds.CreateProjectPath(); err != nil {
		log.Fatal("can't create project path")
	}

	// Project Path to add config path
	projectPath, err := cmds.GetProjectPath()
	if err != nil {
		log.Fatal("project path fetching failed")
	}

	// Logger File
	loggerFile, err := os.OpenFile(filepath.Join(projectPath, "logger.log"), os.O_CREATE|os.O_RDWR|os.O_APPEND, 0o600)
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
	os.OpenFile(filepath.Join(projectPath, "config.yml"), os.O_CREATE|os.O_RDWR, 0o600)

	// Viper setup
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

	// Run the app
	app := cmds.NewApp(logger, viper.GetViper())
	cmds.Execute(app)
}
