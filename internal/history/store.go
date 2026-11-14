package history

import (
	"encoding/csv"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"
	"uuid"
)

type Store struct {
	path string
	mu   sync.Mutex
}

func NewStore(path string) *Store {
	return &Store{path: path}
}

func (s *Store) Append(completedAt time.Time, duration time.Duration, projectID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return fmt.Errorf("project ID is required")
	}

	file, err := os.OpenFile(s.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	defer file.Close()

	recordID := uuid.New()

	csvWriter := csv.NewWriter(file)
	if err := csvWriter.Write([]string{completedAt.Format(time.RFC3339Nano), duration.String(), projectID, recordID.String()}); err != nil {
		return err
	}
	csvWriter.Flush()

	if err := csvWriter.Error(); err != nil {
		return err
	}

	return file.Close()
}

func (s *Store) Load() ([]Record, error) {
	rows, err := readAllCsv(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []Record{}, nil
		}
		return []Record{}, err
	}

	if err := migrate(s.path, rows); err != nil {
		return []Record{}, err
	}

	records := make([]Record, 0, len(rows))
	for index, row := range rows {
		if len(row) != 4 {
			return []Record{}, fmt.Errorf("read session row %d: expected 4 fields, got %d", index+1, len(row))
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
		projectID = strings.TrimSpace(row[2])
		if projectID == "" {
			return []Record{}, fmt.Errorf("read session row %d: project ID is required", index+1)
		}

		recordID, err := uuid.Parse(row[3])
		if err != nil {
			return []Record{}, fmt.Errorf("cant parse record id: %v", err)
		}

		records = append(records, Record{CompletedAt: completedAt, Duration: duration, ProjectID: projectID, RecordID: recordID})
	}

	return records, nil
}

func cleanupFile(file *os.File) {
	file.Close()

	if removeErr := os.Remove(file.Name()); removeErr != nil {
		slog.Error("cant remove file", "remove error", removeErr, "path", file.Name())
	}
}

func migrate(path string, rows [][]string) error {
	clonedRows := slices.Clone(rows)

	migrated := false

	for ind, row := range rows {
		switch len(row) {
		case 2:
			rows[ind] = []string{row[0], row[1], "Unassigned", uuid.New().String()}
			migrated = true
		case 3:
			rows[ind] = []string{row[0], row[1], row[2], uuid.New().String()}
			migrated = true
		}
	}

	if !migrated {
		return nil
	}

	for _, row := range rows {
		if len(row) != 4 {
			return errors.New("migration failed due to missing fields.")
		}

		if _, err := time.Parse(time.RFC3339Nano, row[0]); err != nil {
			return fmt.Errorf("created at parsing failed: %w", err)
		}

		if _, err := time.ParseDuration(row[1]); err != nil {
			return fmt.Errorf("duration cant be parsed: %w", err)
		}

		if strings.TrimSpace(row[2]) == "" {
			return errors.New("project should be set")
		}

		if _, err := uuid.Parse(row[3]); err != nil {
			return fmt.Errorf("uuid cant be parsed: %w", err)
		}
	}

	sessionsDir, _ := filepath.Split(path)

	if _, err := createBackup("sessions-migration", sessionsDir, clonedRows); err != nil {
		return fmt.Errorf("cant create sessions backup: %w", err)
	}

	tmpFile, err := createTemp("migrate-tmp", sessionsDir, rows)
	if err != nil {
		return fmt.Errorf("cant create migration template file: %w", err)
	}

	if err := os.Rename(tmpFile.Name(), path); err != nil {
		cleanupFile(tmpFile)
		return fmt.Errorf("cant over write tmp file: %w", err)
	}

	return nil
}

func writeCsv(file *os.File, rows [][]string) error {
	writer := csv.NewWriter(file)
	if err := writer.WriteAll(rows); err != nil {
		return fmt.Errorf("cant write: %w", err)
	}

	return nil
}

func readAllCsv(path string) ([][]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return [][]string{}, fmt.Errorf("cant open the file: %w", err)
	}
	defer file.Close()

	csvReader := csv.NewReader(file)
	csvReader.FieldsPerRecord = -1
	records, err := csvReader.ReadAll()
	if err != nil {
		return [][]string{}, fmt.Errorf("cant read the sessions: %w", err)
	}

	if err := file.Close(); err != nil {
		return [][]string{}, fmt.Errorf("cant close file: %v", err)
	}

	return records, nil
}

func (s *Store) Delete(record string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	records, err := readAllCsv(s.path)
	if err != nil {
		return err
	}

	newRecords, err := deleteRecord(records, record)
	if err != nil {
		return err
	}

	sessionsDir, _ := filepath.Split(s.path)

	if _, err := createBackup("sessions-delete", sessionsDir, records); err != nil {
		return fmt.Errorf("cant create sessions backup: %w", err)
	}

	var tmpFile *os.File
	if tmpFile, err = createTemp("sessions-tmp", sessionsDir, newRecords); err != nil {
		return err
	}

	err = os.Rename(tmpFile.Name(), s.path)
	if err != nil {
		cleanupFile(tmpFile)
		return fmt.Errorf("cant do renaming: %w", err)
	}

	return nil
}

func deleteRecord(records [][]string, deleteRecord string) ([][]string, error) {
	clonedRecords := slices.Clone(records)

	newRecords := [][]string{}
	deleteRecordExists := false
	for _, r := range clonedRecords {
		if r[3] == deleteRecord {
			deleteRecordExists = true
			continue
		}

		newRecords = append(newRecords, r)
	}

	if !deleteRecordExists {
		return nil, errors.New("delete record does not exist")
	}

	return newRecords, nil
}

func createTemp(prefix, path string, records [][]string) (*os.File, error) {
	clonedRecords := slices.Clone(records)

	file, err := os.CreateTemp(path, fmt.Sprintf("%s-*.csv", prefix))
	if err != nil {
		return nil, fmt.Errorf("can't create tmp file: %w", err)
	}
	defer file.Close()

	if err := writeCsv(file, clonedRecords); err != nil {
		cleanupFile(file)
		return nil, fmt.Errorf("cant write to tmp file: %w", err)
	}

	if err := file.Close(); err != nil {
		cleanupFile(file)
		return nil, fmt.Errorf("cant close the file: %w", err)
	}

	return file, nil
}

func createBackup(prefix, path string, records [][]string) (string, error) {
	filename := filepath.Join(
		path,
		fmt.Sprintf("%s-%s.csv", prefix, time.Now().Format("2006-01-02-15-04-05.000")),
	)

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0o600)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if err := writeCsv(file, records); err != nil {
		cleanupFile(file)
		return "", err
	}

	if err := file.Close(); err != nil {
		cleanupFile(file)
		return "", err
	}

	return filename, nil
}

func (s *Store) TodaysTotal(records []Record) time.Duration {
	total := time.Duration(0)
	for _, record := range records {
		if record.CompletedAt.Format("2006-01-02") == time.Now().Format("2006-01-02") {
			total += record.Duration
		}
	}

	return total
}

func (s *Store) ThisWeek(records []Record) time.Duration {
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

func (s *Store) ThisMonth(records []Record) time.Duration {
	total := time.Duration(0)
	for _, record := range records {
		if time.Now().Month() == record.CompletedAt.Month() && time.Now().Year() == record.CompletedAt.Year() {
			total += record.Duration
		}
	}

	return total
}

func (s *Store) AllTime(records []Record) time.Duration {
	total := time.Duration(0)
	for _, record := range records {
		total += record.Duration
	}

	return total
}
