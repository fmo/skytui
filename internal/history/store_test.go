package history

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"reflect"
	"strings"
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

	if len(rows[0]) != 4 || rows[0][2] != "project-1" {
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
	if records[0].ProjectID != "Unassigned" {
		t.Fatalf("got project ID %q, want Unassigned", records[0].ProjectID)
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
	if records[0].ProjectID != "Unassigned" {
		t.Fatalf("legacy record has project ID %q, want Unassigned", records[0].ProjectID)
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

func TestDeleteRecordDoesNotExist(t *testing.T) {
	now := time.Now()
	store := NewStore(filepath.Join(t.TempDir(), "sessions.csv"))
	store.Append(now, time.Minute, "project")
	records, err := store.Load()
	if err != nil {
		t.Errorf("cant get records: %v", err)
	}
	if len(records) != 1 {
		t.Errorf("records should be nil")
	}

	records, err = store.Load()
	if err != nil {
		t.Errorf("cant load records")
	}

	err = store.Delete("notexistingrecord")
	if err == nil || !strings.Contains(err.Error(), "delete record does not exist") {
		t.Error("error should be there")
	}
}

func TestDeleteCreatesBackupOfOriginalFile(t *testing.T) {
	now := time.Now()

	dir := t.TempDir()

	store := NewStore(filepath.Join(dir, "sessions.csv"))

	store.Append(now, time.Minute, "project")
	store.Append(now.Add(-1*time.Minute), time.Minute*10, "project")
	store.Append(now.Add(-1*time.Hour), time.Hour*2, "project")
	store.Append(now.Add(-3*time.Minute), time.Minute*1, "project")

	records, err := store.Load()
	if err != nil {
		t.Errorf("cant load records: %v", err)
	}

	err = store.Delete(records[0].RecordID.String())
	if err != nil {
		t.Errorf("error occurred: %v", err)
	}

	dirs, err := os.ReadDir(dir)
	if err != nil {
		t.Error("cant read the directory")
	}

	var bupFile os.DirEntry
	for _, d := range dirs {
		if d.Name() == "sessions.csv" {
			continue
		}
		bupFile = d
	}

	buf, err := os.Open(filepath.Join(dir, bupFile.Name()))
	if err != nil {
		t.Error("cant open tmp file")
	}

	csvReader := csv.NewReader(buf)
	rows, err := csvReader.ReadAll()
	buf.Close()
	if err != nil {
		t.Error("cant read records")
	}

	if len(rows) != 4 {
		t.Errorf("there has to be two records but there is %d", len(rows))
	}

	if rows[0][0] != now.Format(time.RFC3339Nano) {
		t.Errorf("order is not expected, record: %v, got: %v", rows[0][0], now.Format(time.RFC3339Nano))
	}

	if rows[1][1] != (time.Minute * 10).String() {
		t.Errorf("order is not expected, record: %v, got: %v", rows[0][1], (time.Minute * 10).String())
	}

	if rows[3][0] != now.Add(-3*time.Minute).Format(time.RFC3339Nano) {
		t.Errorf("order is not expected, records: %v", records)
	}

}

func TestReplaceFile(t *testing.T) {
	dir := t.TempDir()

	record1Time := time.Now()

	store := NewStore(filepath.Join(dir, "sessions.csv"))
	store.Append(record1Time, time.Minute, "project-1")
	store.Append(time.Now().Add(-5*time.Minute), time.Minute, "project-2")

	records, err := store.Load()
	if err != nil {
		t.Errorf("cant load records: %v", err)
	}

	err = store.Delete(records[0].RecordID.String())
	if err != nil {
		t.Errorf("cant delete the record: %v", err)
	}

	records, err = store.Load()
	if err != nil {
		t.Errorf("cant load records: %v", err)
	}

	if len(records) != 1 {
		t.Errorf("there should be one record but there are %d", len(records))
	}
}

func TestDeleteWithMissingField(t *testing.T) {
	dir := t.TempDir()

	projectTime := time.Now().Add(-time.Minute * 2)

	row := time.Now().Format(time.RFC3339Nano) + ",20m\n"
	row += time.Now().Add(-time.Minute*2).Format(time.RFC3339Nano) + ",20m\n"
	row += projectTime.Format(time.RFC3339Nano) + ",30m0s,project-1\n"

	err := os.WriteFile(filepath.Join(dir, "sessions.csv"), []byte(row), 0o600)
	if err != nil {
		t.Error("cant create the file")
	}

	store := NewStore(filepath.Join(dir, "sessions.csv"))
	records, err := store.Load()
	if err != nil {
		t.Error("cant load the records")
	}
	if len(records) != 3 {
		t.Errorf("expected records: 3, got: %d", len(records))
	}
	if err := store.Delete(records[2].RecordID.String()); err != nil {
		t.Errorf("cant delete the project-1: %v", err)
	}
	records, err = store.Load()
	if err != nil {
		t.Errorf("cant load records")
	}
	if len(records) != 2 {
		t.Errorf("there has to be two records but there is %d", len(records))
	}
}

func TestDeleteOrder(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "sessions.csv")
	now := time.Now().Format(time.RFC3339Nano)

	row := now + ",25m,p1\n"
	row += now + ",10m,p1\n"
	row += now + ",5m,p1\n"
	row += now + ",6m10s,p1\n"

	if err := os.WriteFile(f, []byte(row), 0o600); err != nil {
		t.Errorf("cant write to file: %v", err)
	}

	store := NewStore(f)

	records, err := store.Load()
	if err != nil {
		t.Errorf("cant load records")
	}

	if err := store.Delete(records[2].RecordID.String()); err != nil {
		t.Errorf("cant delete records due to: %v", err)
	}
	records, err = store.Load()
	if err != nil {
		t.Errorf("cant load records: %v", err)
	}

	if records[0].Duration.String() != "25m0s" {
		t.Errorf("expected: %v, got: %v", "25m0s", records[0].Duration.String())
	}

	if records[2].Duration.String() != "6m10s" {
		t.Errorf("expected: %v, got: %v", "6m10s", records[2].Duration.String())
	}
}

func TestCreateBackupReturnsSavedPath(t *testing.T) {
	dir := t.TempDir()
	want := [][]string{
		{"2026-11-12T10:00:00Z", "25m0s", "project-1", "550e8400-e29b-41d4-a716-446655440000"},
		{"2026-11-12T11:00:00Z", "10m0s", "project-2", "550e8400-e29b-41d4-a716-446655440001"},
	}

	path, err := createBackup("sessions-manual", dir, want)
	if err != nil {
		t.Fatalf("create backup: %v", err)
	}
	if path == "" {
		t.Fatal("createBackup returned an empty path")
	}
	if filepath.Dir(path) != filepath.Clean(dir) {
		t.Fatalf("backup directory = %q, want %q", filepath.Dir(path), dir)
	}

	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("open returned backup path %q: %v", path, err)
	}
	defer file.Close()

	got, err := csv.NewReader(file).ReadAll()
	if err != nil {
		t.Fatalf("read backup: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("backup records = %#v, want %#v", got, want)
	}
}

func TestWriteCsvFail(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "sessions.csv")

	if err := os.WriteFile(filename, nil, 0o600); err != nil {
		t.Fatalf("create test file: %v", err)
	}

	// Open provides a read-only handle on every supported platform.
	file, err := os.Open(filename)
	if err != nil {
		t.Fatalf("open read-only test file: %v", err)
	}
	defer file.Close()

	err = writeCsv(file, [][]string{{"", "22m0s", "project", ""}})
	if err == nil {
		t.Error("writeCsv succeeded with a read-only file; want a write error")
	}
}

func TestInvalidCsvFail(t *testing.T) {
	const validLegacyRow = "2026-11-11T12:00:00Z,22m0s,project1\n"
	tests := []struct {
		name    string
		rows    string
		wantErr string
	}{
		{
			name:    "two-field timestamp",
			rows:    "invalid-date,22m0s\n",
			wantErr: "created at parsing failed",
		},
		{
			name:    "three-field timestamp",
			rows:    "invalid-date,22m0s,project1\n",
			wantErr: "created at parsing failed",
		},
		{
			name:    "duration",
			rows:    "2026-11-11T12:00:00Z,invalid-duration,project1\n",
			wantErr: "duration cant be parsed",
		},
		{
			name:    "blank project",
			rows:    "2026-11-11T12:00:00Z,22m0s,   \n",
			wantErr: "project should be set",
		},
		{
			name:    "invalid UUID after valid legacy row",
			rows:    validLegacyRow + "2026-11-11T13:00:00Z,25m0s,project1,invalid-uuid\n",
			wantErr: "uuid cant be parsed",
		},
		{
			name:    "unsupported field count after valid legacy row",
			rows:    validLegacyRow + "2026-11-11T13:00:00Z,25m0s,project1,extra,field\n",
			wantErr: "missing fields",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			filename := filepath.Join(dir, "sessions.csv")
			if err := os.WriteFile(filename, []byte(tt.rows), 0o600); err != nil {
				t.Fatalf("write original history: %v", err)
			}

			_, err := NewStore(filename).Load()
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Load error = %v, want error containing %q", err, tt.wantErr)
			}

			contents, err := os.ReadFile(filename)
			if err != nil {
				t.Fatalf("read original history: %v", err)
			}
			if string(contents) != tt.rows {
				t.Errorf("invalid history was modified: got %q, want %q", contents, tt.rows)
			}

			entries, err := os.ReadDir(dir)
			if err != nil {
				t.Fatalf("list history directory: %v", err)
			}
			if len(entries) != 1 || entries[0].Name() != "sessions.csv" {
				t.Errorf("validation failure created extra files: %v", entries)
			}
		})
	}
}
