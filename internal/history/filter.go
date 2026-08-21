package history

func Filter(records []Record, projectID string) []Record {
	filtered := []Record{}

	for _, record := range records {
		if record.ProjectID != projectID {
			continue
		}
		filtered = append(filtered, record)
	}

	return filtered
}
