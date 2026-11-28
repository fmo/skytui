package history

import (
	"encoding/csv"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"testing/synctest"
	"time"
)

func TestUnassignProjectPreservesSessionDataAndOrder(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sessions.csv")
	original := strings.Join([]string{
		"2026-11-27T12:00:00.123456789+03:00,25m0s,project-1,11111111-1111-4111-8111-111111111111",
		"2026-11-26T09:00:00Z,50m0s,project-10,22222222-2222-4222-8222-222222222222",
		"2026-11-25T08:00:00Z,10m0s,Unassigned,33333333-3333-4333-8333-333333333333",
		"2026-11-27T13:00:00+03:00,15m30s,project-1,44444444-4444-4444-8444-444444444444",
	}, "\n") + "\n"
	if err := os.WriteFile(path, []byte(original), 0o600); err != nil {
		t.Fatalf("write original history: %v", err)
	}

	store := NewStore(path)
	if err := store.UnassignProject("project-1"); err != nil {
		t.Fatalf("unassign project sessions: %v", err)
	}

	want := strings.Join([]string{
		"2026-11-27T12:00:00.123456789+03:00,25m0s,Unassigned,11111111-1111-4111-8111-111111111111",
		"2026-11-26T09:00:00Z,50m0s,project-10,22222222-2222-4222-8222-222222222222",
		"2026-11-25T08:00:00Z,10m0s,Unassigned,33333333-3333-4333-8333-333333333333",
		"2026-11-27T13:00:00+03:00,15m30s,Unassigned,44444444-4444-4444-8444-444444444444",
	}, "\n") + "\n"
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read updated history: %v", err)
	}
	if string(got) != want {
		t.Fatalf("reassignment should change only matching project IDs, preserving every other field and row order:\ngot:\n%swant:\n%s", got, want)
	}
	backups, err := filepath.Glob(filepath.Join(filepath.Dir(path), "sessions-project-delete-*.csv"))
	if err != nil || len(backups) != 1 {
		t.Fatalf("expected one history backup, got %v (error: %v)", backups, err)
	}
	backup, err := os.ReadFile(backups[0])
	if err != nil {
		t.Fatalf("read history backup: %v", err)
	}
	if string(backup) != original {
		t.Fatalf("backup should preserve history before reassignment:\ngot:\n%swant:\n%s", backup, original)
	}
}

func TestUnassignProjectMatchesLoadedProjectIDs(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sessions.csv")
	const original = "2026-11-27T12:00:00Z,25m0s, project-1 ,11111111-1111-4111-8111-111111111111\n"
	if err := os.WriteFile(path, []byte(original), 0o600); err != nil {
		t.Fatalf("write history: %v", err)
	}
	store := NewStore(path)
	before, err := store.LoadForProject("project-1")
	if err != nil || len(before) != 1 {
		t.Fatalf("loader should recognize the padded project ID: %#v (error: %v)", before, err)
	}
	if err := store.UnassignProject("project-1"); err != nil {
		t.Fatalf("unassign project: %v", err)
	}
	after, err := store.Load()
	if err != nil || len(after) != 1 {
		t.Fatalf("load reassigned history: %#v (error: %v)", after, err)
	}
	want := before[0]
	want.ProjectID = "Unassigned"
	if after[0] != want {
		t.Fatalf("reassignment should match the loader's project IDs: got %#v, want %#v", after[0], want)
	}
}

func TestUnassignProjectBackupFailurePreservesHistory(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "sessions.csv")
		original := "2026-11-27T12:00:00Z,25m0s,project-1,11111111-1111-4111-8111-111111111111\n"
		if err := os.WriteFile(path, []byte(original), 0o600); err != nil {
			t.Fatalf("write original history: %v", err)
		}
		// The frozen clock makes the backup collide without relying on permissions.
		backupPath := filepath.Join(dir, "sessions-project-delete-"+time.Now().Format("2006-01-02-15-04-05.000")+".csv")
		const existingBackup = "existing recovery backup\n"
		if err := os.WriteFile(backupPath, []byte(existingBackup), 0o600); err != nil {
			t.Fatalf("create conflicting backup: %v", err)
		}

		if err := NewStore(path).UnassignProject("project-1"); !errors.Is(err, os.ErrExist) {
			t.Fatalf("got error %v, want backup collision", err)
		}
		for filename, want := range map[string]string{path: original, backupPath: existingBackup} {
			got, err := os.ReadFile(filename)
			if err != nil {
				t.Fatalf("read file after backup failure: %v", err)
			}
			if string(got) != want {
				t.Errorf("backup failure changed %s: got %q, want %q", filename, got, want)
			}
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatalf("list history directory: %v", err)
		}
		if len(entries) != 2 {
			t.Fatalf("backup failure should leave only the original history and backup, got %v", entries)
		}
	})
}

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

func TestManualBackupPreservesHistory(t *testing.T) {
	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "sessions.csv")
	const original = "2026-11-12T10:00:00Z,25m0s,project-1,550e8400-e29b-41d4-a716-446655440000\n" +
		"2026-11-12T11:00:00Z,10m0s,project-2,550e8400-e29b-41d4-a716-446655440001\n"
	if err := os.WriteFile(sourcePath, []byte(original), 0o600); err != nil {
		t.Fatalf("write original history: %v", err)
	}

	backupPath, err := NewStore(sourcePath).Backup()
	if err != nil {
		t.Fatalf("back up history: %v", err)
	}
	if backupPath == "" || filepath.Clean(backupPath) == sourcePath {
		t.Fatalf("backup path = %q, want a separate backup file", backupPath)
	}
	if filepath.Dir(backupPath) != dir {
		t.Fatalf("backup directory = %q, want %q", filepath.Dir(backupPath), dir)
	}
	if !strings.HasPrefix(filepath.Base(backupPath), "sessions-manual-") {
		t.Errorf("backup filename = %q, want sessions-manual- prefix", filepath.Base(backupPath))
	}

	backup, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("read backup: %v", err)
	}
	if string(backup) != original {
		t.Errorf("backup contents = %q, want both original records in order: %q", backup, original)
	}

	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("read original history: %v", err)
	}
	if string(source) != original {
		t.Errorf("original history changed: got %q, want %q", source, original)
	}
}

func TestManualBackupWithoutHistory(t *testing.T) {
	tests := []struct {
		name         string
		createSource bool
	}{
		{name: "missing file"},
		{name: "empty file", createSource: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			sourcePath := filepath.Join(dir, "sessions.csv")
			if tt.createSource {
				if err := os.WriteFile(sourcePath, nil, 0o600); err != nil {
					t.Fatalf("create empty history: %v", err)
				}
			}

			backupPath, err := NewStore(sourcePath).Backup()
			if err != nil {
				t.Errorf("Backup error = %v, want nil for history without records", err)
			}
			if backupPath != "" {
				t.Errorf("backup path = %q, want empty", backupPath)
			}

			entries, err := os.ReadDir(dir)
			if err != nil {
				t.Fatalf("list history directory: %v", err)
			}
			var wantNames []string
			if tt.createSource {
				wantNames = []string{"sessions.csv"}
			}
			var gotNames []string
			for _, entry := range entries {
				gotNames = append(gotNames, entry.Name())
			}
			if !reflect.DeepEqual(gotNames, wantNames) {
				t.Errorf("files after backup = %v, want %v", gotNames, wantNames)
			}

			if tt.createSource {
				contents, err := os.ReadFile(sourcePath)
				if err != nil {
					t.Fatalf("read original history: %v", err)
				}
				if len(contents) != 0 {
					t.Errorf("empty history changed: got %q", contents)
				}
			}
		})
	}
}

func TestManualBackupReturnsReadError(t *testing.T) {
	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "sessions.csv")
	const original = "2026-11-12T10:00:00Z,25m0s,project-1,550e8400-e29b-41d4-a716-446655440000\n" +
		"\"unterminated field"
	if err := os.WriteFile(sourcePath, []byte(original), 0o600); err != nil {
		t.Fatalf("write malformed history: %v", err)
	}

	backupPath, err := NewStore(sourcePath).Backup()
	if !errors.Is(err, csv.ErrQuote) {
		t.Errorf("Backup error = %v, want CSV quoting error", err)
	}
	if backupPath != "" {
		t.Errorf("backup path = %q, want empty on read failure", backupPath)
	}

	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("read original history: %v", err)
	}
	if string(source) != original {
		t.Errorf("original history changed: got %q, want %q", source, original)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("list history directory: %v", err)
	}
	if len(entries) != 1 || entries[0].Name() != "sessions.csv" {
		t.Errorf("read failure created extra files: %v", entries)
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

func TestCreateBackupReturnsCreationError(t *testing.T) {
	dir := t.TempDir()
	missingDir := filepath.Join(dir, "missing")
	records := [][]string{
		{"2026-11-12T10:00:00Z", "25m0s", "project-1", "550e8400-e29b-41d4-a716-446655440000"},
	}

	backupPath, err := createBackup("sessions-manual", missingDir, records)
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("createBackup error = %v, want missing-directory error", err)
	}
	if backupPath != "" {
		t.Errorf("backup path = %q, want empty on creation failure", backupPath)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("list parent directory: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("creation failure left files or directories behind: %v", entries)
	}
}

func TestCreateBackupDoesNotOverwriteExistingFile(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		dir := t.TempDir()
		originalRecords := [][]string{
			{"2026-11-12T10:00:00Z", "25m0s", "project-1", "550e8400-e29b-41d4-a716-446655440000"},
			{"2026-11-12T11:00:00Z", "10m0s", "project-2", "550e8400-e29b-41d4-a716-446655440001"},
		}
		originalPath, err := createBackup("sessions-manual", dir, originalRecords)
		if err != nil {
			t.Fatalf("create original backup: %v", err)
		}
		original, err := os.ReadFile(originalPath)
		if err != nil {
			t.Fatalf("read original backup: %v", err)
		}

		// The bubble's clock stays fixed, so both requests use the same filename.
		replacementRecords := [][]string{
			{"2026-11-12T12:00:00Z", "5m0s", "project-3", "550e8400-e29b-41d4-a716-446655440002"},
		}
		backupPath, err := createBackup("sessions-manual", dir, replacementRecords)
		if !errors.Is(err, os.ErrExist) {
			t.Errorf("createBackup error = %v, want existing-file error", err)
		}
		if backupPath != "" {
			t.Errorf("backup path = %q, want empty on collision", backupPath)
		}

		contents, err := os.ReadFile(originalPath)
		if err != nil {
			t.Fatalf("read backup after collision: %v", err)
		}
		if string(contents) != string(original) {
			t.Errorf("collision changed existing backup: got %q, want %q", contents, original)
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatalf("list backup directory: %v", err)
		}
		if len(entries) != 1 || entries[0].Name() != filepath.Base(originalPath) {
			t.Errorf("collision changed directory contents: %v", entries)
		}
	})
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
