package BO

import (
	"github.com/LL-res/CRM/common/key"
)

func (m Metric) NoModelKey() key.NoModelKey {
	return key.NoModelKey{
		MetricName:  m.Name,
		MetricUnit:  m.Unit,
		MetricQuery: m.Query,
	}
}

func (m Metric) WithModelKey(modelType string) key.WithModelKey {
	return key.WithModelKey{
		MetricName:  m.Name,
		MetricUnit:  m.Unit,
		MetricQuery: m.Query,
		ModelType:   modelType,
	}
}

//func (m Metric) WithModelKey(modelType string) key.WithModelKey {
//	return fmt.Sprintf("%s$%s", m.NoModelKey(), modelType)
//}
