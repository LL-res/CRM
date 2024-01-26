package collector

import (
	"github.com/LL-res/CRM/collector/collectorType"
	"github.com/LL-res/CRM/collector/prometheusCollector"
	"github.com/LL-res/CRM/common/consts"
	"github.com/LL-res/CRM/common/errs"
	"github.com/LL-res/CRM/common/key"
	"time"
)

type CollectorFacade interface {
	Init(config collectorType.CollectorConfig) error
	CreateCollector(nmk key.NoModelKey)
	DeleteCollector(nmk key.NoModelKey)
	GetMetricFromCollector(nmk key.NoModelKey, length int) []float64
	GetCapFromCollector(nmk key.NoModelKey) int
	GetCollectorKeySet() map[key.NoModelKey]struct{}
	WaitToGetMetric(nmk key.NoModelKey, length int) []float64
}

func GetOrNewFacade(scrapeIntervalDuration time.Duration, historyLength int, url, collectorType string, maxCap int) (CollectorFacade, error) {
	switch collectorType {
	case consts.PROMETHEUS:
		return prometheusCollector.GetOrNewFacade(scrapeIntervalDuration, historyLength, url, maxCap)
	}
	return nil, errs.UNDEFINED_COLLECTOR
}

func GetFacade(collectorType string) CollectorFacade {
	switch collectorType {
	case consts.PROMETHEUS:
		return prometheusCollector.GetFacade()
	default:
		return prometheusCollector.GetFacade()
	}
}
