package apppaths

import (
	"path/filepath"
	"testing"
)

func TestResolveDarwinPaths(t *testing.T) {
	homePath := t.TempDir()
	appPath := filepath.Join(homePath, "Library", "Application Support", "skytui")
	want := Paths{
		ConfigFile:   filepath.Join(appPath, "config.yaml"),
		ProjectsFile: filepath.Join(appPath, "projects.csv"),
		SessionsFile: filepath.Join(appPath, "sessions.csv"),
		LogFile:      filepath.Join(homePath, "Library", "Logs", "skytui", "skytui.log"),
	}

	got, err := resolve("darwin", homePath, emptyEnvironment)
	if err != nil {
		t.Fatalf("resolve Darwin paths: %v", err)
	}
	if got != want {
		t.Fatalf("got paths %#v, want %#v", got, want)
	}
}

func TestResolveLinuxDefaultPaths(t *testing.T) {
	homePath := t.TempDir()
	want := Paths{
		ConfigFile:   filepath.Join(homePath, ".config", "skytui", "config.yaml"),
		ProjectsFile: filepath.Join(homePath, ".local", "share", "skytui", "projects.csv"),
		SessionsFile: filepath.Join(homePath, ".local", "share", "skytui", "sessions.csv"),
		LogFile:      filepath.Join(homePath, ".local", "state", "skytui", "skytui.log"),
	}

	got, err := resolve("linux", homePath, emptyEnvironment)
	if err != nil {
		t.Fatalf("resolve Linux paths: %v", err)
	}
	if got != want {
		t.Fatalf("got paths %#v, want %#v", got, want)
	}
}

func TestResolveLinuxXDGPaths(t *testing.T) {
	basePath := t.TempDir()
	environment := mapEnvironment(map[string]string{
		"XDG_CONFIG_HOME": filepath.Join(basePath, "config"),
		"XDG_DATA_HOME":   filepath.Join(basePath, "data"),
		"XDG_STATE_HOME":  filepath.Join(basePath, "state"),
	})
	want := Paths{
		ConfigFile:   filepath.Join(basePath, "config", "skytui", "config.yaml"),
		ProjectsFile: filepath.Join(basePath, "data", "skytui", "projects.csv"),
		SessionsFile: filepath.Join(basePath, "data", "skytui", "sessions.csv"),
		LogFile:      filepath.Join(basePath, "state", "skytui", "skytui.log"),
	}

	got, err := resolve("linux", basePath, environment)
	if err != nil {
		t.Fatalf("resolve Linux XDG paths: %v", err)
	}
	if got != want {
		t.Fatalf("got paths %#v, want %#v", got, want)
	}
}

func TestResolveLinuxIgnoresRelativeXDGPaths(t *testing.T) {
	homePath := t.TempDir()
	environment := mapEnvironment(map[string]string{
		"XDG_CONFIG_HOME": "config",
		"XDG_DATA_HOME":   "data",
		"XDG_STATE_HOME":  "state",
	})

	got, err := resolve("linux", homePath, environment)
	if err != nil {
		t.Fatalf("resolve Linux paths: %v", err)
	}
	if got.ConfigFile != filepath.Join(homePath, ".config", "skytui", "config.yaml") ||
		got.ProjectsFile != filepath.Join(homePath, ".local", "share", "skytui", "projects.csv") ||
		got.SessionsFile != filepath.Join(homePath, ".local", "share", "skytui", "sessions.csv") ||
		got.LogFile != filepath.Join(homePath, ".local", "state", "skytui", "skytui.log") {
		t.Fatalf("relative XDG paths should use fallbacks, got %#v", got)
	}
}

func TestResolveWindowsEnvironmentPaths(t *testing.T) {
	basePath := t.TempDir()
	roamingPath := filepath.Join(basePath, "roaming")
	localPath := filepath.Join(basePath, "local")
	environment := mapEnvironment(map[string]string{
		"APPDATA":      roamingPath,
		"LOCALAPPDATA": localPath,
	})
	want := Paths{
		ConfigFile:   filepath.Join(roamingPath, "skytui", "config.yaml"),
		ProjectsFile: filepath.Join(localPath, "skytui", "projects.csv"),
		SessionsFile: filepath.Join(localPath, "skytui", "sessions.csv"),
		LogFile:      filepath.Join(localPath, "skytui", "logs", "skytui.log"),
	}

	got, err := resolve("windows", basePath, environment)
	if err != nil {
		t.Fatalf("resolve Windows paths: %v", err)
	}
	if got != want {
		t.Fatalf("got paths %#v, want %#v", got, want)
	}
}

func TestResolveWindowsFallbackPaths(t *testing.T) {
	homePath := t.TempDir()
	roamingPath := filepath.Join(homePath, "AppData", "Roaming")
	localPath := filepath.Join(homePath, "AppData", "Local")
	want := Paths{
		ConfigFile:   filepath.Join(roamingPath, "skytui", "config.yaml"),
		ProjectsFile: filepath.Join(localPath, "skytui", "projects.csv"),
		SessionsFile: filepath.Join(localPath, "skytui", "sessions.csv"),
		LogFile:      filepath.Join(localPath, "skytui", "logs", "skytui.log"),
	}

	got, err := resolve("windows", homePath, emptyEnvironment)
	if err != nil {
		t.Fatalf("resolve Windows fallback paths: %v", err)
	}
	if got != want {
		t.Fatalf("got paths %#v, want %#v", got, want)
	}
}

func TestResolveRejectsUnsupportedOperatingSystem(t *testing.T) {
	_, err := resolve("plan9", t.TempDir(), emptyEnvironment)
	if err == nil {
		t.Fatal("unsupported operating system should return an error")
	}
}

func emptyEnvironment(string) string {
	return ""
}

func mapEnvironment(values map[string]string) func(string) string {
	return func(key string) string {
		return values[key]
	}
}
