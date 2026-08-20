package history

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
	"time"
)

type Store struct {
	path string
}

func NewStore(path string) Store {
	return Store{path: path}
}

func (s Store) Append(completedAt time.Time, duration time.Duration, projectID string) error {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return fmt.Errorf("project ID is required")
	}

	file, err := os.OpenFile(s.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	defer file.Close()

	csvWriter := csv.NewWriter(file)
	if err := csvWriter.Write([]string{completedAt.Format(time.RFC3339Nano), duration.String(), projectID}); err != nil {
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
		if os.IsNotExist(err) {
			return []Record{}, nil
		}
		return []Record{}, err
	}
	defer file.Close()

	csvReader := csv.NewReader(file)
	csvReader.FieldsPerRecord = -1
	rows, err := csvReader.ReadAll()
	if err != nil {
		return []Record{}, err
	}

	records := make([]Record, 0, len(rows))
	for index, row := range rows {
		if len(row) != 2 && len(row) != 3 {
			return []Record{}, fmt.Errorf("read session row %d: expected 2 or 3 fields, got %d", index+1, len(row))
		}
		completedAt, err := time.Parse(time.RFC3339Nano, row[0])
		if err != nil {
			return []Record{}, fmt.Errorf("read session row %d completion time: %w", index+1, err)
		}
		duration, err := time.ParseDuration(row[1])
		if err != nil {
			return []Record{}, fmt.Errorf("read session row %d duration: %w", index+1, err)
		}

		projectID := ""
		if len(row) == 3 {
			projectID = strings.TrimSpace(row[2])
			if projectID == "" {
				return []Record{}, fmt.Errorf("read session row %d: project ID is required", index+1)
			}
		}

		records = append(records, Record{CompletedAt: completedAt, Duration: duration, ProjectID: projectID})
	}

	return records, nil
}

func (s Store) TodaysTotal(records []Record) time.Duration {
	total := time.Duration(0)
	for _, record := range records {
		if record.CompletedAt.Format("2006-01-02") == time.Now().Format("2006-01-02") {
			total += record.Duration
		}
	}

	return total
}

func (s Store) ThisWeek(records []Record) time.Duration {
	thisWeek := time.Duration(0)
	for _, record := range records {
		currentYear, currentWeek := time.Now().ISOWeek()
		completeAtYear, completeAtWeek := record.CompletedAt.ISOWeek()
		if currentWeek == completeAtWeek && currentYear == completeAtYear {
			thisWeek += record.Duration
		}
	}

	return thisWeek
}

func (s Store) ThisMonth(records []Record) time.Duration {
	total := time.Duration(0)
	for _, record := range records {
		if time.Now().Month() == record.CompletedAt.Month() && time.Now().Year() == record.CompletedAt.Year() {
			total += record.Duration
		}
	}

	return total
}

func (s Store) AllTime(records []Record) time.Duration {
	total := time.Duration(0)
	for _, record := range records {
		total += record.Duration
	}

	return total
}
