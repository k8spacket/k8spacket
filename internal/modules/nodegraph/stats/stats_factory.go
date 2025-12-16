package stats

type StatsFactory struct {
}

func (factory *StatsFactory) GetStats(statsType string) Stats {
	switch statsType {
	case "bytes":
		return &BytesStats{}
	case "duration":
		return &DurationStats{}
	default:
		return &ConnectionStats{}
	}
}
