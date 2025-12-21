package storer

import (
	"testing"
	"time"

	"github.com/k8spacket/k8spacket/internal/modules/tlsparser/model"
	"github.com/k8spacket/k8spacket/internal/modules/tlsparser/repository"
	"github.com/k8spacket/k8spacket/internal/modules/tlsparser/update"
	"github.com/stretchr/testify/assert"
)

type mockRepository struct {
	repository.Repository
	resultConnection model.TLSConnection
	resultDetails    model.TLSDetails
}

func (mockRepository *mockRepository) Query(from time.Time, to time.Time) []model.TLSConnection {
	return []model.TLSConnection{}
}

func (mockRepository *mockRepository) UpsertConnection(key string, value *model.TLSConnection) {
	mockRepository.resultConnection = *value
}

func (mockRepository *mockRepository) Read(key string) model.TLSDetails {
	return model.TLSDetails{}
}

func (mockRepository *mockRepository) UpsertDetails(key string, value *model.TLSDetails, fn repository.Fn) {
	fn(value, &mockRepository.resultDetails)
	mockRepository.resultDetails = *value
}

type mockCertificateUpdater struct {
	update.CertificateUpdater
	fnCalled bool
}

func (mockCertificateUpdater *mockCertificateUpdater) Update(newValue *model.TLSDetails, oldValue *model.TLSDetails) {
	mockCertificateUpdater.fnCalled = true
}

func TestStoreInDatabase(t *testing.T) {

	mockRepository := &mockRepository{}
	mockCertificateUpdater := &mockCertificateUpdater{}

	storer := NewStorer(mockRepository, mockCertificateUpdater)

	tlsConnection := model.TLSConnection{Src: "src"}
	tlsDetails := model.TLSDetails{UsedTLSVersion: "TLS 1.2"}

	storer.StoreInDatabase(&tlsConnection, &tlsDetails)

	assert.NotEmpty(t, mockRepository.resultConnection)
	assert.NotEmpty(t, mockRepository.resultDetails)

	assert.NotEmpty(t, mockRepository.resultConnection.Id)
	assert.NotEmpty(t, mockRepository.resultDetails.Id)

	assert.EqualValues(t, true, mockCertificateUpdater.fnCalled)

}
