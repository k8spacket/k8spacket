package repository

import (
	"github.com/k8spacket/k8spacket/modules/nodegraph/model"
	"regexp"
	"time"
)

type IRepository[T model.ConnectionItem] interface {
	Read(key string) T
	Query(from time.Time, to time.Time, patternNs *regexp.Regexp, patternIn *regexp.Regexp, patternEx *regexp.Regexp) []T
	Set(key string, value *T)
}
