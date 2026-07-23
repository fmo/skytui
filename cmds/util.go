package cmds

import (
	"os"
	"path/filepath"
)

func GetProjectPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, "Library", "Application Support", "skytui"), nil
}

func CreateProjectPath() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	fullpath := filepath.Join(homeDir, "Library", "Application Support", "skytui")

	return os.MkdirAll(fullpath, 0o600)
}
