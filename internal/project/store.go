package project

import (
	"crypto/rand"
	"encoding/csv"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

var (
	ErrNameRequired = errors.New("project name is required")
	ErrNameExists   = errors.New("project name already exists")
	ErrNotFound     = errors.New("project not found")
)

type Store struct {
	path     string
	projects []Project
}

func NewStore(path string) (*Store, error) {
	projects, err := load(path)
	if err != nil {
		return nil, err
	}

	return &Store{path: path, projects: projects}, nil
}

func (s *Store) Create(name string) (Project, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Project{}, ErrNameRequired
	}
	if !isUniqueName(s.projects, name) {
		return Project{}, ErrNameExists
	}

	id, err := newID()
	if err != nil {
		return Project{}, fmt.Errorf("generate project ID: %w", err)
	}
	project := Project{ID: id, Name: name}

	file, err := os.OpenFile(s.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return Project{}, fmt.Errorf("open project store: %w", err)
	}
	defer file.Close()

	csvWriter := csv.NewWriter(file)
	if err := csvWriter.Write([]string{project.ID, project.Name}); err != nil {
		return Project{}, fmt.Errorf("write project: %w", err)
	}
	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		return Project{}, fmt.Errorf("flush project store: %w", err)
	}

	s.projects = append(s.projects, project)
	return project, nil
}

func (s *Store) Delete(projectID string) error {
	projects, err := load(s.path)
	if err != nil {
		return err
	}
	index := slices.IndexFunc(projects, func(p Project) bool { return p.ID == projectID })
	if index < 0 {
		return ErrNotFound
	}
	remaining := slices.Delete(projects, index, index+1)
	rows := make([][]string, 0, len(remaining))
	for _, p := range remaining {
		rows = append(rows, []string{p.ID, p.Name})
	}

	if err := backupProjectFile(s.path); err != nil {
		return fmt.Errorf("back up projects before deletion: %w", err)
	}

	tmp, err := os.CreateTemp(filepath.Dir(s.path), "projects-tmp-*.csv")
	if err != nil {
		return fmt.Errorf("create project replacement: %w", err)
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	if err := csv.NewWriter(tmp).WriteAll(rows); err != nil {
		return fmt.Errorf("write project replacement: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close project replacement: %w", err)
	}
	if err := os.Rename(tmp.Name(), s.path); err != nil {
		return fmt.Errorf("replace project store: %w", err)
	}
	s.projects = remaining
	return nil
}

func backupProjectFile(path string) error {
	contents, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read original project file: %w", err)
	}
	backupPath := filepath.Join(filepath.Dir(path), "projects-delete-"+time.Now().Format("2006-01-02-15-04-05.000")+".csv")
	file, err := os.OpenFile(backupPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	complete := false
	defer func() {
		file.Close()
		if !complete {
			if err := os.Remove(backupPath); err != nil {
				slog.Error("remove incomplete project backup", "path", backupPath, "err", err)
			}
		}
	}()
	if _, err := file.Write(contents); err != nil {
		return fmt.Errorf("write project backup: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close project backup: %w", err)
	}
	complete = true
	return nil
}

func (s *Store) List() []Project {
	return slices.Clone(s.projects)
}

func load(path string) ([]Project, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Project{}, nil
		}
		return nil, fmt.Errorf("open project store: %w", err)
	}
	defer file.Close()

	rows, err := csv.NewReader(file).ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read project store: %w", err)
	}

	projects := make([]Project, 0, len(rows))
	ids := make(map[string]struct{}, len(rows))
	for index, row := range rows {
		if len(row) != 2 {
			return nil, fmt.Errorf("read project row %d: expected 2 fields, got %d", index+1, len(row))
		}

		project := Project{ID: row[0], Name: strings.TrimSpace(row[1])}
		if project.ID == "" {
			return nil, fmt.Errorf("read project row %d: project ID is required", index+1)
		}
		if project.Name == "" {
			return nil, fmt.Errorf("read project row %d: %w", index+1, ErrNameRequired)
		}
		if _, exists := ids[project.ID]; exists {
			return nil, fmt.Errorf("read project row %d: duplicate project ID %q", index+1, project.ID)
		}
		if !isUniqueName(projects, project.Name) {
			return nil, fmt.Errorf("read project row %d: %w", index+1, ErrNameExists)
		}

		ids[project.ID] = struct{}{}
		projects = append(projects, project)
	}

	return projects, nil
}

func isUniqueName(projects []Project, name string) bool {
	for _, project := range projects {
		if strings.EqualFold(project.Name, name) {
			return false
		}
	}

	return true
}

func newID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}
