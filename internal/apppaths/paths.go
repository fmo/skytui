package apppaths

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

type Paths struct {
	ConfigFile   string
	ProjectsFile string
	SessionsFile string
	LogFile      string
}

func resolve(goos, homePath string, getenv func(string) string) (Paths, error) {
	switch goos {
	case "darwin":
		appPath := filepath.Join(homePath, "Library", "Application Support", "skytui")

		return Paths{
			ConfigFile:   filepath.Join(appPath, "config.yaml"),
			ProjectsFile: filepath.Join(appPath, "projects.csv"),
			SessionsFile: filepath.Join(appPath, "sessions.csv"),
			LogFile:      filepath.Join(homePath, "Library", "Logs", "skytui", "skytui.log"),
		}, nil
	case "linux":
		configBase := xdgDir(
			getenv("XDG_CONFIG_HOME"),
			filepath.Join(homePath, ".config"),
		)
		dataBase := xdgDir(
			getenv("XDG_DATA_HOME"),
			filepath.Join(homePath, ".local", "share"),
		)
		stateBase := xdgDir(
			getenv("XDG_STATE_HOME"),
			filepath.Join(homePath, ".local", "state"),
		)

		return Paths{
			ConfigFile: filepath.Join(
				configBase,
				"skytui",
				"config.yaml",
			),
			ProjectsFile: filepath.Join(
				dataBase,
				"skytui",
				"projects.csv",
			),
			SessionsFile: filepath.Join(
				dataBase,
				"skytui",
				"sessions.csv",
			),
			LogFile: filepath.Join(
				stateBase,
				"skytui",
				"skytui.log",
			),
		}, nil
	case "windows":
		configBase := getenv("APPDATA")
		if configBase == "" {
			configBase = filepath.Join(homePath, "AppData", "Roaming")
		}

		localBase := getenv("LOCALAPPDATA")
		if localBase == "" {
			localBase = filepath.Join(homePath, "AppData", "Local")
		}

		return Paths{
			ConfigFile: filepath.Join(
				configBase,
				"skytui",
				"config.yaml",
			),
			ProjectsFile: filepath.Join(
				localBase,
				"skytui",
				"projects.csv",
			),
			SessionsFile: filepath.Join(
				localBase,
				"skytui",
				"sessions.csv",
			),
			LogFile: filepath.Join(
				localBase,
				"skytui",
				"logs",
				"skytui.log",
			),
		}, nil
	default:
		return Paths{}, fmt.Errorf("unsupported operating system: %s", goos)
	}
}

func Resolve() (Paths, error) {
	homePath, err := os.UserHomeDir()
	if err != nil {
		return Paths{}, fmt.Errorf("get user home directory: %w", err)
	}

	return resolve(runtime.GOOS, homePath, os.Getenv)
}

func xdgDir(value, fallback string) string {
	if value == "" || !filepath.IsAbs(value) {
		return fallback
	}

	return value
}
