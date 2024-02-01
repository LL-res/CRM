package prometheusCollector

import (
	"github.com/LL-res/CRM/collector/collectorType"
	"github.com/LL-res/CRM/common/key"
	"github.com/LL-res/CRM/common/log"
	"github.com/prometheus/client_golang/api"
	"strings"
	"time"
)

type PrometheusCollectorFacade struct {
	collectors     map[key.NoModelKey]*PrometheusCollector
	stopChans      map[key.NoModelKey]chan struct{}
	maxCap         int
	scrapeInterval time.Duration
	//prometheus client
	client        api.Client
	historyLength int
}

func (p *PrometheusCollectorFacade) GetCollectorKeySet() map[key.NoModelKey]struct{} {
	result := make(map[key.NoModelKey]struct{})
	for k := range p.collectors {
		result[k] = struct{}{}
	}
	return result
}

var facade *PrometheusCollectorFacade

func GetOrNewFacade(scrapeInterval time.Duration, historyLength int, url string, maxCap int) (*PrometheusCollectorFacade, error) {
	if facade == nil {
		facade = &PrometheusCollectorFacade{
			collectors:     make(map[key.NoModelKey]*PrometheusCollector, 0),
			stopChans:      make(map[key.NoModelKey]chan struct{}, 0),
			scrapeInterval: scrapeInterval,
			historyLength:  historyLength,
			maxCap:         maxCap,
		}
		err := facade.setServerAddress(url)
		if err != nil {
			log.Logger.Error(err, "")
			return nil, err
		}
	}
	return facade, nil
}

func GetFacade() *PrometheusCollectorFacade {
	return facade
}

func (p *PrometheusCollectorFacade) Init(config collectorType.CollectorConfig) error {
	if strings.HasSuffix(config.DataSource, "exp") {
		return nil
	}
	client, err := api.NewClient(api.Config{
		Address: config.DataSource,
	})
	if err != nil {
		return err
	}
	p.client = client

	return nil
}

func (p *PrometheusCollectorFacade) CreateCollector(nmk key.NoModelKey) {
	pCollector := NewPrometheusCollector(p.client, nmk, p.scrapeInterval, p.historyLength, p.maxCap)
	p.collectors[nmk] = pCollector
	stopChan := make(chan struct{}, 1)
	go pCollector.Collect(stopChan)
	p.stopChans[nmk] = stopChan
}

func (p *PrometheusCollectorFacade) DeleteCollector(nmk key.NoModelKey) {
	stopChan := p.stopChans[nmk]
	stopChan <- struct{}{}
	delete(p.collectors, nmk)
}

func (p *PrometheusCollectorFacade) GetMetricFromCollector(nmk key.NoModelKey, length int) []float64 {
	targetCollector := p.collectors[nmk]
	targetCollector.GetMetrics(length)
	return targetCollector.GetMetrics(length)
}

func (p *PrometheusCollectorFacade) GetCapFromCollector(nmk key.NoModelKey) int {
	targetCollector := p.collectors[nmk]
	return targetCollector.DataCap()
}

func (p *PrometheusCollectorFacade) setServerAddress(url string) error {
	if strings.HasSuffix(url, "exp") {
		return nil
	}
	client, err := api.NewClient(api.Config{
		Address: url,
	})
	if err != nil {
		return err
	}
	p.client = client

	return nil
}
func (p *PrometheusCollectorFacade) WaitToGetMetric(nmk key.NoModelKey, length int) []float64 {
	targetCollector := p.collectors[nmk]
	for ; ; <-time.Tick(time.Second) {
		if targetCollector.DataCap() < length {
			continue
		}
		return targetCollector.GetMetrics(length)
	}
}
