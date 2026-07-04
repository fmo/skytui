package cmds

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type PomodoroManager struct {
	pomodoroFile *os.File
}

func NewPomodoroManager(pomodoroFile *os.File) *PomodoroManager {
	return &PomodoroManager{pomodoroFile}
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
