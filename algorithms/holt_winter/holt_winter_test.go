package holt_winter

import (
	"context"
	"fmt"
	mock_collector "github.com/LL-res/CRM/collector/mockCollector"
	"github.com/LL-res/CRM/common/key"
	"github.com/LL-res/CRM/domain/BO"
	"github.com/LL-res/CRM/predictor/message"
	"github.com/golang/mock/gomock"
	"testing"
)

var modelAttr map[string]string
var metrics []float64

func setup() {
	modelAttr = make(map[string]string)
	modelAttr["slen"] = "12"
	modelAttr["alpha"] = "0.716"
	modelAttr["beta"] = "0.029"
	modelAttr["gamma"] = "0.993"
	default_base := []float64{30, 21, 29, 31, 40, 48, 53, 47, 37, 39, 31, 29, 17, 9, 20, 24, 27, 35, 41, 38,
		27, 31, 27, 26, 21, 13, 21, 18, 33, 35, 40, 36, 22, 24, 21, 20, 17, 14, 17, 19,
		26, 29, 40, 31, 20, 24, 18, 26, 17, 9, 17, 21, 28, 32, 46, 33, 23, 28, 22, 27,
		18, 8, 17, 21, 31, 34, 44, 38, 31, 30, 26, 32}
	for i := 0; i < 2000; i++ {
		metrics = append(metrics, default_base[i%len(default_base)])
	}

}

func TestMain(m *testing.M) {
	setup()
	m.Run()
}

func TestPredict(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	facade := mock_collector.NewMockCollectorFacade(ctl)
	facade.EXPECT().GetMetricFromCollector(gomock.Any(), gomock.Any()).Return(metrics)
	facade.EXPECT().GetCapFromCollector(gomock.Any()).Return(200)
	pred, _ := New(message.Param{
		WithModelKey: key.WithModelKey{
			MetricName:  "name",
			MetricUnit:  "unit",
			MetricQuery: "query",
			ModelType:   "type",
		},
		CollectorFacade: facade,
		LookForward:     24,
		Model: BO.Model{
			LookBackward: 60,
			Debug:        true,
			Attr:         modelAttr,
		},
	})
	_, err := pred.Predict(context.Background())
	if err != nil {
		t.Errorf(err.Error())
	}
}

// LookBackward: 200,LookForward: 120 效果较好
func TestLoss(t *testing.T) {
	lookForward := 60
	lookBackward := 120
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	facade := mock_collector.NewMockCollectorFacade(ctl)
	facade.EXPECT().GetMetricFromCollector(gomock.Any(), gomock.Any()).Return(metrics[len(metrics)-lookBackward-lookForward:])
	facade.EXPECT().GetCapFromCollector(gomock.Any()).Return(200)
	pred, _ := New(message.Param{
		WithModelKey: key.WithModelKey{
			MetricName:  "name",
			MetricUnit:  "unit",
			MetricQuery: "query",
			ModelType:   "type",
		},
		CollectorFacade: facade,
		LookForward:     lookForward,
		Model: BO.Model{
			LookBackward: lookBackward,
			Debug:        true,
			Attr:         modelAttr,
		},
	})
	loss, err := pred.Loss()
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	fmt.Println(loss)
}
