package certificate

import (
	"github.com/k8spacket/k8spacket/modules/tls-parser/model"
)

type ICertificate interface {
	UpdateCertificateInfo(newValue *model.TLSDetails, oldValue *model.TLSDetails)
}
