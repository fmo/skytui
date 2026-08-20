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
	if err := store.Append(time.Now(), time.Second*10); err != nil {
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
}

func TestLoad(t *testing.T) {
	testFile := filepath.Join(t.TempDir(), "sessions.csv")

	completedAt := time.Now().Add(-10 * time.Second)
	duration := time.Second * 20

	store := NewStore(testFile)

	err := store.Append(completedAt, duration)
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
}

func TestTodaysTotal(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sessions.csv")

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, now.Location())
	yesterday := today.AddDate(0, 0, -1)

	store := NewStore(path)

	if err := store.Append(today, time.Second*10); err != nil {
		t.Fatalf("cant append the record: %v", err)
	}
	if err := store.Append(today.Add(-10*time.Minute), time.Minute*3); err != nil {
		t.Fatalf("cant append the record: %v", err)
	}
	if err := store.Append(yesterday, time.Minute*3); err != nil {
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
