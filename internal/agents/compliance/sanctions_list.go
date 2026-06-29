package compliance

// InMemorySanctionsList implements SanctionsList with an in-memory list of entries.
// This can be used for testing or as a fallback data source.
type InMemorySanctionsList struct {
	entries []string
}

// NewInMemorySanctionsList creates a new InMemorySanctionsList with the given entries.
func NewInMemorySanctionsList(entries []string) *InMemorySanctionsList {
	return &InMemorySanctionsList{entries: entries}
}

// GetEntries returns all entries in the sanctions list.
func (l *InMemorySanctionsList) GetEntries() []string {
	return l.entries
}

// SampleSanctionsList returns a sample OFAC sanctions list for testing/demonstration.
func SampleSanctionsList() *InMemorySanctionsList {
	return NewInMemorySanctionsList([]string{
		"Osama Bin Laden",
		"Al-Qaeda Foundation",
		"Hezbollah Financial Services",
		"Kim Jong Un",
		"Islamic Revolutionary Guard Corps",
		"Banco Delta Asia",
		"Viktor Bout",
		"Sinaloa Cartel Trading Company",
		"Russian Military Intelligence GRU",
		"Wagner Group PMC",
	})
}
