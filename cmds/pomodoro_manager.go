package cmds

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type PomodoroManager struct {
	pomodoroFile *os.File
	logger       *slog.Logger
}

func NewPomodoroManager(pomodoroFile *os.File, logger *slog.Logger) *PomodoroManager {
	return &PomodoroManager{pomodoroFile, logger}
}

func (m *PomodoroManager) Save(limit, count int) error {
	if m.pomodoroFile == nil {
		return errors.New("pomodoro file is nil")
	}

	w := csv.NewWriter(m.pomodoroFile)

	// completed is total duration in seconds - count down time, ie total 30 seconds - 10 count down (left)
	completed := time.Duration(limit-count) * time.Second

	left := strconv.Itoa(count)

	// current time, completed time, count down
	err := w.Write([]string{time.Now().Format(time.RFC3339), completed.String(), fmt.Sprintf("%ss", left)})
	if err != nil {
		return err
	}
	w.Flush()

	return nil
}

func (m *PomodoroManager) RenameFile(oldFile, newFile string) error {
	projectPath, err := GetProjectPath(false)
	if err != nil {
		return err
	}

	oldFileFull := filepath.Join(projectPath, oldFile)
	newFileFull := filepath.Join(projectPath, newFile)

	err = os.Rename(oldFileFull, newFileFull)
	if err != nil {
		return err
	}

	return nil
}

func (m *PomodoroManager) CreatePomodoroBackup(backupFile string) error {
	src, err := os.Open(m.pomodoroFile.Name())
	if err != nil {
		return err
	}

	fp, err := GetProjectPath(false)
	if err != nil {
		return err
	}

	dsc, err := os.Create(filepath.Join(fp, backupFile))
	if err != nil {
		return err
	}

	io.Copy(dsc, src)

	return nil
}

func (m *PomodoroManager) Restore(pomodoroFilename, backupFilename string) error {
	projectPath, err := GetProjectPath(false)
	if err != nil {
		return err
	}

	backupFile, err := os.Open(filepath.Join(projectPath, backupFilename))
	if err != nil {
		return err
	}

	pomodoroFile, err := os.Create(filepath.Join(projectPath, pomodoroFilename))
	if err != nil {
		return err
	}

	_, err = io.Copy(pomodoroFile, backupFile)
	if err != nil {
		m.logger.Error("cant copy", "err", err)
		return err
	}

	return nil
}
