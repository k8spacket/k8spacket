package stats

type IFactory interface {
	GetStats(statsType string) IStats
}
