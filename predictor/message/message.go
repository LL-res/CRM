package message

import (
	"github.com/LL-res/CRM/collector"
	"github.com/LL-res/CRM/common/key"
	"github.com/LL-res/CRM/domain/BO"
	"time"
)

type PredictResult struct {
	StartMetric     float64
	Loss            float64
	PredictMetric   []float64 `json:"predictMetric"`
	LastPointOfReal time.Time
}

type PredictRequest struct {
	Metrics []float64        `json:"metrics"`
	Key     key.WithModelKey `json:"key"`
	//由于模型未保存参数，故需要参数恢复模型
	//模型的超参数，每个模型都是特定的
	ModelAttr map[string]string `json:"modelAttr"`
	//模型参数中的公共部分
	LookForward  int `json:"lookForward"`
	LookBackward int `json:"lookBackward"`
}

type TrainRequest struct {
	Metrics []float64        `json:"metrics"`
	Key     key.WithModelKey `json:"key"`
	//用于构建模型
	ModelAttr    map[string]string `json:"modelAttr"`
	LookForward  int               `json:"lookForward"`
	LookBackward int               `json:"lookBackward"`
}

type Param struct {
	WithModelKey    key.WithModelKey
	CollectorFacade collector.CollectorFacade
	//所有模型都只有一个统一的lookForward
	LookForward int
	//模型的配置各个模型可以不同，其中包含一些公共的参数配置，与每个模型特有的超参数
	Model BO.Model
}
