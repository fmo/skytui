package session

import (
	"encoding/csv"
	"os"
	"time"
)

type Store struct {
	path string
}

func New(path string) Store {
	return Store{path: path}
}

func (s Store) Append(completedAt time.Time, duration time.Duration) error {
	file, err := os.OpenFile(s.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.FileMode(0o600))
	if err != nil {
		return err
	}
	defer file.Close()

	csvWriter := csv.NewWriter(file)
	if err := csvWriter.Write([]string{completedAt.Format(time.RFC3339Nano), duration.String()}); err != nil {
		return err
	}
	csvWriter.Flush()

	if err := csvWriter.Error(); err != nil {
		return err
	}

	return nil
}
