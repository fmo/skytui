package cmds

import (
	"os"
	"path/filepath"
)

func GetProjectPath(createDirs bool) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	projectPath := filepath.Join(home, "Library", "Application Support", "skytui")

	if createDirs {
		if err := os.MkdirAll(projectPath, 0o700); err != nil {
			return "", err
		}
	}

	return projectPath, nil
}

func OpenFile(filename string) (*os.File, error) {
	fp, err := GetProjectPath(false)
	if err != nil {
		return nil, err
	}

	fullFileName := filepath.Join(fp, filename)

	var f *os.File

	f, err = os.OpenFile(fullFileName, os.O_APPEND|os.O_RDWR, 0o700)
	if err != nil {
		f, err = os.Create(fullFileName)
		if err != nil {
			return nil, err
		}
	}

	return f, nil
}
