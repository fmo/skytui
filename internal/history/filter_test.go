package history

import (
	"testing"
	"time"
)

func TestFilterRecords(t *testing.T) {
	records := []Record{
		{Duration: 10 * time.Minute, ProjectID: ""},
		{Duration: 20 * time.Minute, ProjectID: "project-1"},
		{Duration: 30 * time.Minute, ProjectID: "project-10"},
	}

	tests := []struct {
		name    string
		filter  Filter
		wantIDs []string
	}{
		{
			name:    "all projects",
			filter:  Filter{Mode: AllProjects},
			wantIDs: []string{"", "project-1", "project-10"},
		},
		{
			name:    "one exact project",
			filter:  Filter{Mode: OneProject, ProjectID: "project-1"},
			wantIDs: []string{"project-1"},
		},
		{
			name:    "unassigned",
			filter:  Filter{Mode: Unassigned},
			wantIDs: []string{""},
		},
		{
			name:    "project without records",
			filter:  Filter{Mode: OneProject, ProjectID: "missing"},
			wantIDs: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := FilterRecords(records, tt.filter)
			if len(filtered) != len(tt.wantIDs) {
				t.Fatalf("got %d records, want %d", len(filtered), len(tt.wantIDs))
			}
			for index, wantID := range tt.wantIDs {
				if filtered[index].ProjectID != wantID {
					t.Fatalf("record %d has project ID %q, want %q", index, filtered[index].ProjectID, wantID)
				}
			}
		})
	}
}

func TestAllProjectsFilterReturnsCopy(t *testing.T) {
	records := []Record{{ProjectID: "project-1"}}
	filtered := FilterRecords(records, Filter{Mode: AllProjects})
	filtered[0].ProjectID = "changed"

	if records[0].ProjectID != "project-1" {
		t.Fatal("all-project filtering changed the source records")
	}
}
