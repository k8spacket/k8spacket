package stats

type Factory interface {
	GetStats(statsType string) Stats
}
