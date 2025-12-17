package stats

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetStats(t *testing.T) {
	t.Parallel()
	var tests = []struct {
		statsType string
		want      Stats
	}{
		{"bytes", &BytesStats{}},
		{"duration", &DurationStats{}},
		{"connection", &ConnectionStats{}},
		{"", &ConnectionStats{}},
	}

	factory := StatsFactory{}

	for _, test := range tests {
		t.Run(test.statsType, func(t *testing.T) {
			t.Parallel()
			result := factory.GetStats(test.statsType)
			assert.Equal(t, reflect.TypeOf(test.want).String(), reflect.TypeOf(result).String())
		})
	}
}
