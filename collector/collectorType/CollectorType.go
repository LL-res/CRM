package collectorType

import "time"

type CollectorConfig struct {
	DataSource     string
	ScrapeInterval time.Duration
	MaxCap         int
}

type Metric struct {
	Value     float64
	TimeStamp time.Time
}
