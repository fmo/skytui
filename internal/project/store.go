package project

import (
	"crypto/rand"
	"encoding/csv"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
)

var (
	ErrNameRequired = errors.New("project name is required")
	ErrNameExists   = errors.New("project name already exists")
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
