package compliance

// DebarmentList is an interface for the federal debarment data source,
// enabling mocking in tests.
type DebarmentList interface {
	// GetEntries returns all entries in the debarment list.
	GetEntries() []string
}

// InMemoryDebarmentList implements DebarmentList with an in-memory list of entries.
// This can be used for testing or as a fallback data source.
type InMemoryDebarmentList struct {
	entries []string
}

// NewInMemoryDebarmentList creates a new InMemoryDebarmentList with the given entries.
func NewInMemoryDebarmentList(entries []string) *InMemoryDebarmentList {
	return &InMemoryDebarmentList{entries: entries}
}

// GetEntries returns all entries in the debarment list.
func (l *InMemoryDebarmentList) GetEntries() []string {
	return l.entries
}

// SampleDebarmentList returns a sample federal debarment list for testing/demonstration.
func SampleDebarmentList() *InMemoryDebarmentList {
	return NewInMemoryDebarmentList([]string{
		"Blackwater Security LLC",
		"Continental Defense Systems",
		"Pacific Rim Contractors Inc",
		"GlobalTech Solutions Group",
		"Apex Military Services",
		"Meridian Construction Corp",
		"Atlas Federal Consulting",
		"Pinnacle Defense Industries",
		"Frontier Logistics International",
		"Sovereign Engineering Partners",
	})
}
