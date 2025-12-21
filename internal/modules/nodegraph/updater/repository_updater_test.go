package updater

import (
	"github.com/k8spacket/k8spacket/internal/modules/nodegraph/model"
	"github.com/k8spacket/k8spacket/internal/modules/nodegraph/repository"
	"github.com/stretchr/testify/assert"
	"testing"
)

type mockRepository struct {
	repository.Repository[model.ConnectionItem]
	result model.ConnectionItem
}

func (mock *mockRepository) Set(key string, value *model.ConnectionItem) {
	mock.result = *value
}

func (mock *mockRepository) Read(key string) model.ConnectionItem {
	return mock.result
}

func TestUpdate(t *testing.T) {
	var tests = []struct {
		item model.ConnectionItem
		want model.ConnectionItem
	}{
		{model.ConnectionItem{Src: "src", Dst: "dst", ConnCount: 10, ConnPersistent: 5, BytesReceived: 1000, BytesSent: 500, Duration: 0.5, MaxDuration: 0.5},
			model.ConnectionItem{Src: "src", SrcName: "srcName", SrcNamespace: "srcNs", Dst: "dst", DstName: "dstName", DstNamespace: "dstNs", ConnCount: 11, ConnPersistent: 6, BytesSent: 600, BytesReceived: 1200, Duration: 1.5, MaxDuration: 1}},
		{model.ConnectionItem{},
			model.ConnectionItem{Src: "src", SrcName: "srcName", SrcNamespace: "srcNs", Dst: "dst", DstName: "dstName", DstNamespace: "dstNs", ConnCount: 1, ConnPersistent: 1, BytesSent: 100, BytesReceived: 200, Duration: 1, MaxDuration: 1}},
	}

	for _, test := range tests {
		t.Run(test.item.Src, func(t *testing.T) {

			mockRepository := &mockRepository{result: test.item}
			updater := NewUpdater(mockRepository)

			updater.Update("src", "srcName", "srcNs", "dst", "dstName", "dstNs", true, 100, 200, 1, true)

			result := mockRepository.Read("")

			test.want.LastSeen = result.LastSeen
			assert.EqualValues(t, test.want, result)
		})
	}
}
