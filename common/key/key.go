package key

type WithModelKey struct {
	MetricName  string `json:"metricName"`
	MetricUnit  string `json:"metricUnit"`
	MetricQuery string `json:"metricQuery"`
	ModelType   string `json:"modelType"`
}

func (wmk WithModelKey) IsEmpty() bool {
	return wmk.MetricName == "" && wmk.MetricUnit == "" && wmk.MetricQuery == "" && wmk.ModelType == ""
}

func (wmk WithModelKey) ToNoModelKey() NoModelKey {
	return NoModelKey{
		MetricName:  wmk.MetricName,
		MetricUnit:  wmk.MetricUnit,
		MetricQuery: wmk.MetricQuery,
	}
}

type NoModelKey struct {
	MetricName  string `json:"metricName"`
	MetricUnit  string `json:"metricUnit"`
	MetricQuery string `json:"metricQuery"`
}

func (nmk NoModelKey) IsEmpty() bool {
	return nmk.MetricName == "" && nmk.MetricUnit == "" && nmk.MetricQuery == ""
}

func (nmk NoModelKey) ToWithModelKey(modelType string) WithModelKey {
	return WithModelKey{
		MetricName:  nmk.MetricName,
		MetricUnit:  nmk.MetricUnit,
		MetricQuery: nmk.MetricQuery,
		ModelType:   modelType,
	}
}
