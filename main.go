package main

import (
	"log"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/fmo/skytui/cmd"
	"github.com/fmo/skytui/internal/apppaths"
	"github.com/fmo/skytui/internal/config"
	"github.com/fmo/skytui/internal/history"
	"github.com/fmo/skytui/internal/notifier"
	"github.com/fmo/skytui/internal/project"
)

func main() {
	paths, err := apppaths.Resolve()
	if err != nil {
		log.Fatalf("cant resolve application paths: %v", err)
	}

	// logger setup
	if err := os.MkdirAll(filepath.Dir(paths.LogFile), 0o700); err != nil {
		log.Fatalf("cant create logs directory: %v", err)
	}

	logFile, err := os.OpenFile(paths.LogFile, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0o600)
	if err != nil {
		log.Fatal("cant open log file")
	}
	defer logFile.Close()

	logHandler := slog.NewTextHandler(logFile, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	})
	logger := slog.New(logHandler)

	logger.Info("starting SkyTUI application")

	slog.SetDefault(logger)

	for _, filePath := range []string{paths.ConfigFile, paths.ProjectsFile, paths.SessionsFile} {
		if err := os.MkdirAll(filepath.Dir(filePath), 0o700); err != nil {
			logger.Error("cant create application directory", "path", filepath.Dir(filePath), "err", err)
			os.Exit(1)
		}
	}

	cfg, err := config.New(filepath.Dir(paths.ConfigFile))
	if err != nil {
		logger.Error("cant load configuration", "err", err)
		os.Exit(1)
	}

	defaultFocusDuration, err := cfg.LoadDefaultFocusDuration()
	if err != nil {
		logger.Error("cant load configuration", "err", err)
		os.Exit(1)
	}

	shortBreakDuration, err := cfg.LoadShortBreakDuration()
	if err != nil {
		logger.Error("cant load configuration", "err", err)
		os.Exit(1)
	}
	notificationsEnabled := cfg.LoadNotificationsEnabled()

	projectStore, err := project.NewStore(paths.ProjectsFile)
	if err != nil {
		logger.Error("cant get project store", "err", err)
		os.Exit(1)
	}

	// history store
	historyStore := history.NewStore(paths.SessionsFile)
	if err := cmd.Exec(historyStore, projectStore, cfg, defaultFocusDuration, shortBreakDuration, notificationsEnabled, notifier.New()); err != nil {
		logFile.Close()
		log.Fatalf("cant run the command: %v", err)
	}
}
