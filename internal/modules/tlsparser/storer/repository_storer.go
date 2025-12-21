package storer

import (
	"fmt"
	"github.com/k8spacket/k8spacket/internal/modules/tlsparser/model"
	"github.com/k8spacket/k8spacket/internal/modules/tlsparser/repository"
	"github.com/k8spacket/k8spacket/internal/modules/tlsparser/update"
	"github.com/k8spacket/k8spacket/internal/thirdparty/db"
	"strconv"
)

type RepositoryStorer struct {
	repo    repository.Repository
	updater update.Updater
}

func NewStorer(repo repository.Repository, updater update.Updater) Storer {
	return &RepositoryStorer{repo: repo, updater: updater}
}

func (storer *RepositoryStorer) StoreInDatabase(tlsConnection *model.TLSConnection, tlsDetails *model.TLSDetails) {
	var id = strconv.Itoa(int(db.HashId(fmt.Sprintf("%s-%s", tlsConnection.Src, tlsConnection.Dst))))
	tlsConnection.Id = id
	storer.repo.UpsertConnection(id, tlsConnection)
	tlsDetails.Id = id
	storer.repo.UpsertDetails(id, tlsDetails, storer.updater.Update)
}
