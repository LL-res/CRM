package knowledge

import (
	"github.com/LL-res/CRM/collector"
	"github.com/LL-res/CRM/common/key"
	"github.com/LL-res/CRM/domain/BO"
	"github.com/LL-res/CRM/predictor"
	"github.com/LL-res/CRM/scaler"
	"github.com/LL-res/CRM/utils"
	"k8s.io/apimachinery/pkg/types"
	"time"
)

type ClusterKnowledge map[types.NamespacedName]*LocalKnowledge

type LocalKnowledge struct {
	// the max cap of collector,the value is the max train size
	CollectorCap int
	//noModelKey
	//use to close collector dynamically
	//CollectorMap map[string]chan struct{}
	CollectorFacade collector.CollectorFacade
	//noModelKey
	MetricMap utils.ConcurrentMap[key.NoModelKey, *BO.Metric]
	//withModelKey
	PredictorMap utils.ConcurrentMap[key.WithModelKey, predictor.Predictor]
	//withModelKey
	ModelMap utils.ConcurrentMap[key.WithModelKey, *BO.Model]
	//withModelKey
	//store the latest timestamp the model trained
	TrainHistory utils.ConcurrentMap[key.WithModelKey, time.Time]
	//one scaler for one aom instance
	Scaler *scaler.Scaler
}

func (h *LocalKnowledge) Init() {
	h.CollectorCap = 2000
	h.MetricMap.NewConcurrentMap()
	h.PredictorMap.NewConcurrentMap()
	h.ModelMap.NewConcurrentMap()
	h.TrainHistory.NewConcurrentMap()
}
