package history

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestAppend(t *testing.T) {
	testFile := filepath.Join(t.TempDir(), "sessions.csv")
	store := NewStore(testFile)
	if err := store.Append(time.Now(), time.Second*10, "project-1"); err != nil {
		t.Fatalf("cant append the row: %v", err)
	}

	file, err := os.Open(testFile)
	if err != nil {
		t.Fatal("cant open file", "err", err)
	}
	defer file.Close()

	csvReader := csv.NewReader(file)
	rows, err := csvReader.ReadAll()
	if err != nil {
		t.Fatal("cant read csv rows", "err", err)
	}

	if len(rows) != 1 {
		t.Fatalf("got: %d rows, want: 1", len(rows))
	}

	if rows[0][1] != (time.Second * 10).String() {
		t.Error("append does not work")
	}
	if len(rows[0]) != 3 || rows[0][2] != "project-1" {
		t.Fatalf("got row %#v, want project ID as third field", rows[0])
	}
}

func TestAppendRequiresProjectID(t *testing.T) {
	store := NewStore(filepath.Join(t.TempDir(), "sessions.csv"))
	if err := store.Append(time.Now(), 25*time.Minute, ""); err == nil {
		t.Fatal("append without a project ID should fail")
	}
}

func TestLoad(t *testing.T) {
	testFile := filepath.Join(t.TempDir(), "sessions.csv")

	completedAt := time.Now().Add(-10 * time.Second)
	duration := time.Second * 20

	store := NewStore(testFile)

	err := store.Append(completedAt, duration, "project-1")
	if err != nil {
		t.Fatalf("cant append the record: %v", err)
	}

	records, err := store.Load()
	if err != nil {
		t.Fatalf("cant load the sessions: %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("got: %d, want: 1", len(records))
	}

	if !records[0].CompletedAt.Equal(completedAt) {
		t.Errorf("wanted: %v, got: %v", completedAt, records[0].CompletedAt)
	}

	if records[0].Duration != duration {
		t.Errorf("wanted: %v, got: %v", duration, records[0].Duration)
	}
	if records[0].ProjectID != "project-1" {
		t.Errorf("wanted project ID %q, got %q", "project-1", records[0].ProjectID)
	}
}

func TestLoadLegacyRecordWithoutProject(t *testing.T) {
	testFile := filepath.Join(t.TempDir(), "sessions.csv")
	completedAt := time.Date(2026, time.August, 20, 12, 0, 0, 0, time.UTC)
	row := completedAt.Format(time.RFC3339Nano) + ",25m0s\n"
	if err := os.WriteFile(testFile, []byte(row), 0o600); err != nil {
		t.Fatalf("write legacy session: %v", err)
	}

	records, err := NewStore(testFile).Load()
	if err != nil {
		t.Fatalf("load legacy session: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("got %d records, want 1", len(records))
	}
	if records[0].ProjectID != "" {
		t.Fatalf("got project ID %q, want empty", records[0].ProjectID)
	}
}

func TestLoadMixedLegacyAndProjectRecords(t *testing.T) {
	testFile := filepath.Join(t.TempDir(), "sessions.csv")
	firstCompletedAt := time.Date(2026, time.August, 19, 12, 0, 0, 0, time.UTC)
	secondCompletedAt := firstCompletedAt.Add(24 * time.Hour)
	rows := [][]string{
		{firstCompletedAt.Format(time.RFC3339Nano), "25m0s"},
		{secondCompletedAt.Format(time.RFC3339Nano), "30m0s", "project-1"},
	}

	file, err := os.OpenFile(testFile, os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatalf("create session fixture: %v", err)
	}
	writer := csv.NewWriter(file)
	if err := writer.WriteAll(rows); err != nil {
		file.Close()
		t.Fatalf("write session fixture: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close session fixture: %v", err)
	}

	records, err := NewStore(testFile).Load()
	if err != nil {
		t.Fatalf("load mixed sessions: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("got %d records, want 2", len(records))
	}
	if records[0].ProjectID != "" {
		t.Fatalf("legacy record has project ID %q, want empty", records[0].ProjectID)
	}
	if records[1].ProjectID != "project-1" {
		t.Fatalf("project-aware record has project ID %q, want %q", records[1].ProjectID, "project-1")
	}
}

func TestTodaysTotal(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sessions.csv")

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, now.Location())
	yesterday := today.AddDate(0, 0, -1)

	store := NewStore(path)

	if err := store.Append(today, time.Second*10, "project-1"); err != nil {
		t.Fatalf("cant append the record: %v", err)
	}
	if err := store.Append(today.Add(-10*time.Minute), time.Minute*3, "project-1"); err != nil {
		t.Fatalf("cant append the record: %v", err)
	}
	if err := store.Append(yesterday, time.Minute*3, "project-1"); err != nil {
		t.Fatalf("cant append the record: %v", err)
	}

	records, err := store.Load()
	if err != nil {
		t.Fatalf("cant load the records: %v", err)
	}

	want := (time.Minute * 3) + (10 * time.Second)

	if got := store.TodaysTotal(records); got != want {
		t.Errorf("total time is not correct, want: %v, got: %v", want, got)
	}
}

func TestThisWeek(t *testing.T) {
	store := NewStore("")
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, now.Location())
	weekday := int(today.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	weekStart := today.AddDate(0, 0, 1-weekday)

	records := []Record{
		{CompletedAt: weekStart, Duration: 10 * time.Minute},
		{CompletedAt: weekStart.AddDate(0, 0, 6), Duration: 20 * time.Minute},
		{CompletedAt: weekStart.AddDate(0, 0, -1), Duration: 30 * time.Minute},
		{CompletedAt: weekStart.AddDate(0, 0, 7), Duration: 40 * time.Minute},
	}

	want := 30 * time.Minute
	if got := store.ThisWeek(records); got != want {
		t.Errorf("this week's total is incorrect: got %v, want %v", got, want)
	}
}

func TestThisMonth(t *testing.T) {
	store := NewStore("")
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 12, 0, 0, 0, now.Location())

	records := []Record{
		{CompletedAt: monthStart, Duration: 10 * time.Minute},
		{CompletedAt: monthStart.AddDate(0, 1, -1), Duration: 20 * time.Minute},
		{CompletedAt: monthStart.AddDate(0, 0, -1), Duration: 30 * time.Minute},
		{CompletedAt: monthStart.AddDate(0, 1, 0), Duration: 40 * time.Minute},
	}

	want := 30 * time.Minute
	if got := store.ThisMonth(records); got != want {
		t.Errorf("this month's total is incorrect: got %v, want %v", got, want)
	}
}

func TestAllTime(t *testing.T) {
	store := NewStore("")
	records := []Record{
		{Duration: 10 * time.Minute},
		{Duration: 20 * time.Minute},
		{Duration: 30 * time.Minute},
	}

	want := time.Hour
	if got := store.AllTime(records); got != want {
		t.Errorf("all-time total is incorrect: got %v, want %v", got, want)
	}
}
