package predictor

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
	modelAttr["n_layers"] = "1"
	modelAttr["batch_size"] = "10"
	modelAttr["epochs"] = "20"
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

func TestGRUPredict(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	facade := mock_collector.NewMockCollectorFacade(ctl)
	facade.EXPECT().GetMetricFromCollector(gomock.Any(), gomock.Any()).Return(metrics)
	facade.EXPECT().GetCapFromCollector(gomock.Any()).Return(2000)
	pred := NewExternalPredictor(message.Param{
		WithModelKey: key.WithModelKey{
			MetricName:  "name",
			MetricUnit:  "unit",
			MetricQuery: "query",
			ModelType:   "type",
		},
		CollectorFacade: facade,
		LookForward:     120,
		Model: BO.Model{
			LookBackward:  200,
			Debug:         false,
			SourceImplURL: "http://127.0.0.1:5000",
			Command:       "",
			Attr:          modelAttr,
		},
	})
	_, err := pred.Predict(context.Background())
	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestGRUTrain(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	facade := mock_collector.NewMockCollectorFacade(ctl)
	facade.EXPECT().GetMetricFromCollector(gomock.Any(), gomock.Any()).Return(metrics)
	facade.EXPECT().GetCapFromCollector(gomock.Any()).Return(2000)
	pred := NewExternalPredictor(message.Param{
		WithModelKey: key.WithModelKey{
			MetricName:  "name",
			MetricUnit:  "unit",
			MetricQuery: "query",
			ModelType:   "type",
		},
		CollectorFacade: facade,
		LookForward:     120,
		Model: BO.Model{
			LookBackward:  200,
			Debug:         false,
			SourceImplURL: "http://127.0.0.1:5000",
			Command:       "",
			Attr:          modelAttr,
		},
	})
	finishedChan, errChan := pred.Train(context.Background())
	select {
	case <-finishedChan:
		fmt.Println("success")
	case err := <-errChan:
		t.Error(err)
	}
}

func TestLoss(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	facade := mock_collector.NewMockCollectorFacade(ctl)
	//facade.EXPECT().GetMetricFromCollector(gomock.Any(), gomock.Any()).Return(metrics)
	//facade.EXPECT().GetCapFromCollector(gomock.Any()).Return(2000)
	facade.EXPECT().WaitToGetMetric(gomock.Any(), gomock.Any()).Return(metrics)
	pred := NewExternalPredictor(message.Param{
		WithModelKey: key.WithModelKey{
			MetricName:  "name",
			MetricUnit:  "unit",
			MetricQuery: "query",
			ModelType:   "type",
		},
		CollectorFacade: facade,
		LookForward:     120,
		Model: BO.Model{
			LookBackward:  200,
			Debug:         true,
			SourceImplURL: "http://127.0.0.1:5000",
			Command:       "",
			Attr:          modelAttr,
		},
	})
	loss, err := pred.Loss()
	if err != nil {
		t.Error(err.Error())
		return
	}
	fmt.Println(loss)
}
