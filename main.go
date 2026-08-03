package main

import (
	"log"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/fmo/skytui/cmd"
)

func main() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("cant open user home dir")
	}

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

	if err := cmd.Exec(); err != nil {
		logFile.Close()
		log.Fatalf("cant run the command: %v", err)
	}
}
