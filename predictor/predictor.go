package predictor

import (
	"context"
	"github.com/LL-res/CRM/algorithms/holt_winter"
	"github.com/LL-res/CRM/collector"
	"github.com/LL-res/CRM/common/consts"
	"github.com/LL-res/CRM/common/key"
	"github.com/LL-res/CRM/domain/BO"
	"github.com/LL-res/CRM/predictor/message"
)

type Predictor interface {
	GetType() string
	Key() key.WithModelKey
	Predict(ctx context.Context) (message.PredictResult, error)
	// @description 异步开始训练并返回两个容量为1的channel,两个channel中时刻只有一个channel中有元素
	// @return    chan struct{}    成功标志channel
	// @return    chan error    错误channel
	Train(ctx context.Context) (chan struct{}, chan error)
	// @description 同步等待数据充足以进行一次预测，并将预测值与真实值之间的MSE作为误差值返回
	Loss() (float64, error)
}
type InternalPredictor interface {
	Predictor
}
type ExternalPredictor interface {
	Predictor
	StartImplSource() error
	StopImplSource() error
}

type Param struct {
	WithModelKey    key.WithModelKey
	CollectorFacade collector.CollectorFacade
	LookForward     int
	Model           BO.Model
}

func NewPredictor(param message.Param) (Predictor, error) {
	switch param.WithModelKey.ModelType {
	case consts.HOLT_WINTER:
		pred, err := holt_winter.New(param)
		if err != nil {
			return nil, err
		}
		return pred, nil
	default:
		return NewExternalPredictor(param), nil
	}
}
