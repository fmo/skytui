package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func GetCsvFilename() string {
	csvFile := "pomodoro.csv"

	if os.Getenv("csvfile") != "" {
		csvFile = os.Getenv("csvfile")
	}

	return csvFile
}

func GetCsvBackup(csvfile string) string {
	s := strings.Split(csvfile, ".")
	return fmt.Sprintf("%s_bup.csv", s[0])
}

func GetProjectPath(createDirs bool) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	projectPath := filepath.Join(home, "Library", "Application Support", "pomodoro")

	if createDirs {
		if err := os.MkdirAll(projectPath, 0o700); err != nil {
			return "", err
		}
	}

	return projectPath, nil
}

func OpenPomodoroFile() (*os.File, error) {
	fp, err := GetProjectPath(false)
	if err != nil {
		return nil, err
	}

	fullFileName := filepath.Join(fp, GetCsvFilename())

	var f *os.File

	f, err = os.OpenFile(fullFileName, os.O_APPEND|os.O_WRONLY, 0o700)
	if err != nil {
		f, err = os.Create(fullFileName)
		if err != nil {
			return nil, err
		}
	}

	return f, nil
}

func OpenConfigFile() error {
	projectPath, err := GetProjectPath(false)
	if err != nil {
		return err
	}

	configFile := filepath.Join(projectPath, "config.yml")

	_, err = os.Open(configFile)
	if err != nil {
		_, err = os.Create(configFile)
		if err != nil {
			return err
		}
	}

	return nil
}
