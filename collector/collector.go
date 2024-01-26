package collector

type Collector interface {
	Collect(stopChan <-chan struct{})
	DataCap() int
	GetMetrics(length int) []float64
}
