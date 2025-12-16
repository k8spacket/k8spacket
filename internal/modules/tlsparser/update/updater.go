package update

import (
	"github.com/k8spacket/k8spacket/internal/modules/tlsparser/model"
)

type Updater interface {
	Update(newValue *model.TLSDetails, oldValue *model.TLSDetails)
}
