package cmd

import (
	"fmt"
	"io"
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

func GetCsvFile() (*os.File, error) {
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

func CreateBackup(file *os.File) error {
	src, err := os.Open(file.Name())
	if err != nil {
		return err
	}

	fp, err := GetProjectPath(false)
	if err != nil {
		return err
	}
	csvFile := GetCsvBackup(GetCsvFilename())

	dsc, err := os.Create(filepath.Join(fp, csvFile))
	if err != nil {
		return err
	}

	io.Copy(dsc, src)

	return nil
}
