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

func (s Store) Load() ([]Record, error) {
	file, err := os.Open(s.path)
	if err != nil {
		if os.IsExist(err) {
			return []Record{}, nil
		}
		return []Record{}, err
	}

	csvReader := csv.NewReader(file)
	rows, err := csvReader.ReadAll()
	if err != nil {
		return []Record{}, err
	}

	var records []Record
	for _, row := range rows {
		complatedAt, err := time.Parse(time.RFC3339Nano, row[0])
		if err != nil {
			return []Record{}, err
		}
		duration, err := time.ParseDuration(row[1])
		if err != nil {
			return []Record{}, err
		}

		records = append(records, Record{complatedAt, duration})
	}

	return records, nil
}
