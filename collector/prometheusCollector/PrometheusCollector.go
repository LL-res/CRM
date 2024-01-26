package prometheusCollector

import (
	"context"
	"github.com/LL-res/CRM/common/key"
	"github.com/LL-res/CRM/common/log"
	"github.com/LL-res/CRM/utils"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"time"
)

type PrometheusCollector struct {
	promql  string
	metrics []float64
	client  api.Client
	//基于过去的数据量
	historyLength int
	//最大的指标数据容量
	maxCap         int
	scrapeInterval time.Duration
}

func (p *PrometheusCollector) Collect(stopChan <-chan struct{}) {
	v1api := v1.NewAPI(p.client)
	now := time.Now()
	metricsRaw, _, err := v1api.QueryRange(context.Background(), p.promql, v1.Range{
		Start: utils.CaliberateTime(now.Add(-time.Duration(p.historyLength)*p.scrapeInterval), p.scrapeInterval),
		End:   utils.CaliberateTime(now, p.scrapeInterval),
		Step:  p.scrapeInterval,
	})
	if err != nil {
		log.Logger.Error(err, "")
		return
	}
	metrics := metricsRaw.(model.Matrix)
	for _, samples := range metrics {
		for _, sample := range samples.Values {
			p.metrics = append(p.metrics, float64(sample.Value))
		}
	}
	theLastPoint := utils.CaliberateTime(now, p.scrapeInterval)
	for i := 1; ; i++ {
		select {
		case <-stopChan:
			log.Logger.Info("worker stopped")
			return
		default:
			currentPoint := theLastPoint.Add(time.Duration(i) * p.scrapeInterval)
			if currentPoint.After(time.Now()) {
			outerLoop:
				for {
					select {
					case <-time.After(currentPoint.Sub(time.Now())):
						break outerLoop
					}
				}
			}
			result, _, err := v1api.Query(context.Background(), p.promql, currentPoint)
			if err != nil {
				log.Logger.Error(err, "")
				continue
			}
			vector := result.(model.Vector)
			for _, sample := range vector {
				p.metrics = append(p.metrics, float64(sample.Value))
			}
			if len(p.metrics) > p.maxCap {
				p.metrics = p.metrics[len(p.metrics)-p.maxCap:]
			}
		}
	}
}

func (p *PrometheusCollector) DataCap() int {
	return len(p.metrics)
}

func (p *PrometheusCollector) GetMetrics(length int) []float64 {
	res := make([]float64, length)
	copy(res, p.metrics[len(p.metrics)-length:])
	//w.data = make([]collector.Metric, 0)
	return res
}

func NewPrometheusCollector(client api.Client, nmk key.NoModelKey, scrapeInterval time.Duration, historyLength int, maxCap int) *PrometheusCollector {
	return &PrometheusCollector{
		promql:         nmk.MetricQuery,
		metrics:        make([]float64, 0),
		client:         client,
		maxCap:         maxCap,
		historyLength:  historyLength,
		scrapeInterval: scrapeInterval,
	}
}
