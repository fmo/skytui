package history

import (
	"encoding/csv"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"time"
)

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

func writeCsv(file *os.File, rows [][]string) error {
	writer := csv.NewWriter(file)
	if err := writer.WriteAll(rows); err != nil {
		return fmt.Errorf("cant write: %w", err)
	}

	return nil
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

func cleanupFile(file *os.File) {
	file.Close()

	if removeErr := os.Remove(file.Name()); removeErr != nil {
		slog.Error("cant remove file", "remove error", removeErr, "path", file.Name())
	}
}
