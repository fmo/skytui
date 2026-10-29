package history

import (
	"encoding/csv"
	"errors"
	"fmt"
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

	return nil
}

func (s *Store) Load() ([]Record, error) {
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

	if err := migrate(rows, s.path); err != nil {
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

func migrate(rows [][]string, path string) error {
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

	dir, _ := filepath.Split(path)

	// create backup migrate-bup-2006-01-02-15-04-05.000.csv
	filename := fmt.Sprintf("migrate-bup-%s.csv", time.Now().Format("2006-01-02-15-04-05.000"))
	file, err := os.OpenFile(filepath.Join(dir, filename), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return fmt.Errorf("cant open the file: %v", err)
	}

	writer := csv.NewWriter(file)
	writer.WriteAll(clonedRows)
	writer.Flush()

	file.Close()

	if writer.Error() != nil {
		return fmt.Errorf("cant flush the backup: %w", writer.Error())
	}

	// create tmp
	tmpfile, err := os.CreateTemp(dir, "migrate-tmp-*.csv")
	if err != nil {
		return fmt.Errorf("cant open file: %s", err)
	}

	writer = csv.NewWriter(tmpfile)
	writer.WriteAll(rows)
	writer.Flush()

	tmpfile.Close()

	if err != nil {
		return fmt.Errorf("cant flush the tmp: %v", err)
	}

	// replace original with tmp
	if err := os.Rename(tmpfile.Name(), path); err != nil {
		return fmt.Errorf("cant over write tmp file: %w", err)
	}

	return nil
}

func (s *Store) Delete(record int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := os.Open(s.path)
	if err != nil {
		return fmt.Errorf("cant open the file: %v", err)
	}

	csvReader := csv.NewReader(file)
	csvReader.FieldsPerRecord = -1
	records, err := csvReader.ReadAll()
	if err != nil {
		return fmt.Errorf("cant read the sessions: %w", err)
	}

	file.Close()

	newRecords, err := deleteRecord(record, records)
	if err != nil {
		return err
	}

	fp, _ := filepath.Split(s.path)

	if err := createBackup(fp, records); err != nil {
		return err
	}

	var tmpFileName string
	if tmpFileName, err = createTmpWithDeletedRecords(fp, newRecords); err != nil {
		return err
	}

	// write tmp file over sessions.csv
	err = os.Rename(tmpFileName, filepath.Join(fp, "sessions.csv"))
	if err != nil {
		return fmt.Errorf("cant do renaming: %w", err)
	}

	return nil
}

func deleteRecord(deleteRecord int, records [][]string) ([][]string, error) {
	reversedRecords := slices.Clone(records)

	slices.Reverse(reversedRecords)

	newRecords := [][]string{}
	deleteRecordExists := false
	for ind, r := range reversedRecords {
		if ind == deleteRecord {
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

func createTmpWithDeletedRecords(path string, records [][]string) (string, error) {
	reverseRecords := slices.Clone(records)

	slices.Reverse(reverseRecords)

	tmpFile, err := os.CreateTemp(path, "sessions-*.csv")
	if err != nil {
		return "", fmt.Errorf("can't create tmp file: %w", err)
	}
	defer tmpFile.Close()

	csvWriter := csv.NewWriter(tmpFile)
	err = csvWriter.WriteAll(reverseRecords)
	if err != nil {
		return "", fmt.Errorf("cant write to tmp file: %w", err)
	}
	csvWriter.Flush()

	if csvWriter.Error() != nil {
		return "", fmt.Errorf("cant write tmp file: %v", err)
	}

	return tmpFile.Name(), nil
}

func createBackup(path string, records [][]string) error {
	now := time.Now().Format("2006-01-02-15-04-05.000")
	filename := fmt.Sprintf("sessions-%s.csv", now)

	file, err := os.OpenFile(filepath.Join(path, filename), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return err
	}
	defer file.Close()

	csvWriter := csv.NewWriter(file)
	if err := csvWriter.WriteAll(records); err != nil {
		return fmt.Errorf("cant write to tmp file: %w", err)
	}
	csvWriter.Flush()

	if csvWriter.Error() != nil {
		return fmt.Errorf("cant write backup: %v", err)
	}

	return nil
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
