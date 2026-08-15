package main

import (
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/fmo/skytui/cmd"
	"github.com/fmo/skytui/internal/session"
	"github.com/spf13/viper"
)

const defaultPomodoroDuration = 25 * time.Minute

func loadDefaultDuration(appDir string) (time.Duration, error) {
	config := viper.New()
	config.SetConfigName("config")
	config.SetConfigType("yaml")
	config.AddConfigPath(appDir)
	config.SetDefault("default-duration", defaultPomodoroDuration.String())

	if err := config.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			return 0, fmt.Errorf("read config: %w", err)
		}

		config.Set("default-duration", defaultPomodoroDuration.String())
		if err := config.SafeWriteConfigAs(filepath.Join(appDir, "config.yaml")); err != nil {
			return 0, fmt.Errorf("create config: %w", err)
		}
	}

	duration, err := time.ParseDuration(config.GetString("default-duration"))
	if err != nil {
		return 0, fmt.Errorf("parse default-duration: %w", err)
	}
	if duration < time.Second || duration%time.Second != 0 {
		return 0, fmt.Errorf("default-duration should be at least 1 second and use whole seconds: %v", duration)
	}

	return duration, nil
}

func main() {
	// home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("cant open user home dir")
	}

	// logger setup
	logPath := filepath.Join(homeDir, "Library", "Logs", "skytui")

	if err := os.Mkdir(logPath, 0o700); err != nil && !os.IsExist(err) {
		log.Fatalf("cant create logs directory: %v", err)
	}

	logFile, err := os.OpenFile(filepath.Join(logPath, "skytui.log"), os.O_CREATE|os.O_RDWR|os.O_APPEND, 0o600)
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

	// app directory setup
	appDir := filepath.Join(homeDir, "Library", "Application Support", "skytui")
	if err := os.Mkdir(appDir, 0o700); err != nil && !os.IsExist(err) {
		logger.Error("cant create application folder", "err", err)
		os.Exit(1)
	}

	defaultDuration, err := loadDefaultDuration(appDir)
	if err != nil {
		logger.Error("cant load configuration", "err", err)
		os.Exit(1)
	}

	// store
	store := session.New(filepath.Join(appDir, "sessions.csv"))
	if err := cmd.Exec(store, defaultDuration); err != nil {
		logFile.Close()
		log.Fatalf("cant run the command: %v", err)
	}
}
