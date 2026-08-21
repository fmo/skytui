package history

import "slices"

type FilterMode int

const (
	AllProjects FilterMode = iota
	OneProject
	Unassigned
)

type Filter struct {
	Mode      FilterMode
	ProjectID string
}

func FilterRecords(records []Record, filter Filter) []Record {
	if filter.Mode == AllProjects {
		return slices.Clone(records)
	}

	filtered := make([]Record, 0)

	for _, record := range records {
		switch filter.Mode {
		case OneProject:
			if record.ProjectID == filter.ProjectID {
				filtered = append(filtered, record)
			}
		case Unassigned:
			if record.ProjectID == "" {
				filtered = append(filtered, record)
			}
		}
	}

	return filtered
}
