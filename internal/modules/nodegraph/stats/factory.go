package stats

type Factory struct {
}

func (factory *Factory) GetStats(statsType string) IStats {
	switch statsType {
	case "bytes":
		return &Bytes{}
	case "duration":
		return &Duration{}
	default:
		return &Connection{}
	}
}
