package predictor

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/LL-res/CRM/collector"
	"github.com/LL-res/CRM/common/consts"
	"github.com/LL-res/CRM/common/errs"
	"github.com/LL-res/CRM/common/key"
	"github.com/LL-res/CRM/common/log"
	"github.com/LL-res/CRM/domain/BO"
	"github.com/LL-res/CRM/predictor/message"
	"github.com/LL-res/CRM/utils"
	"github.com/go-resty/resty/v2"
	"net/http"
	"os/exec"
)

type ExternalPredictorImpl struct {
	wmk         key.WithModelKey
	facade      collector.CollectorFacade
	lookForward int
	model       BO.Model
}

func (e *ExternalPredictorImpl) StartImplSource() error {
	if e.model.Command == "" {
		log.Logger.Info("No command specified")
		return nil
	}
	cmd := exec.Command(e.model.Command)
	err := cmd.Run()
	if err != nil {
		return errors.New("failed to execute command: " + err.Error())
	}
	return nil
}

func (e *ExternalPredictorImpl) StopImplSource() error {
	return nil
}

func (e *ExternalPredictorImpl) Init() {
	//TODO implement me
	return
}

func (e *ExternalPredictorImpl) GetType() string {
	return e.wmk.ModelType
}

func (e *ExternalPredictorImpl) Key() key.WithModelKey {
	return e.wmk
}

func (e *ExternalPredictorImpl) Predict(ctx context.Context) (message.PredictResult, error) {
	capacityForNow := e.facade.GetCapFromCollector(e.wmk.ToNoModelKey())
	if capacityForNow < e.model.LookBackward {
		return message.PredictResult{}, errs.NO_SUFFICENT_DATA
	}
	client := resty.New()
	response, err := client.R().
		SetBody(message.PredictRequest{
			Metrics:      e.facade.GetMetricFromCollector(e.wmk.ToNoModelKey(), e.model.LookBackward),
			Key:          e.wmk,
			ModelAttr:    e.model.Attr,
			LookForward:  e.lookForward,
			LookBackward: e.model.LookBackward,
		}).
		Post(e.model.SourceImplURL + consts.PREDICT_SUFFIX)
	if err != nil {
		log.Logger.Error(err, "")
		return message.PredictResult{}, err
	}
	if response.StatusCode() != http.StatusOK {
		return message.PredictResult{}, errors.New(string(response.Body()))
	}
	result := message.PredictResult{}
	err = json.Unmarshal(response.Body(), &result)
	if err != nil {
		log.Logger.Error(err, "")
		return message.PredictResult{}, err
	}
	return result, nil
}

func (e *ExternalPredictorImpl) Train(ctx context.Context) (chan struct{}, chan error) {
	asyncErrors := make(chan error, 1)
	finishChan := make(chan struct{}, 1)
	capacityForNow := e.facade.GetCapFromCollector(e.wmk.ToNoModelKey())
	if capacityForNow < e.model.TrainSize {
		asyncErrors <- errs.NO_SUFFICENT_DATA
		return finishChan, asyncErrors
	}
	go func(chan struct{}, chan error) {
		client := resty.New()
		response, err := client.R().
			SetBody(message.TrainRequest{
				Metrics:      e.facade.GetMetricFromCollector(e.wmk.ToNoModelKey(), e.model.TrainSize),
				Key:          e.wmk,
				ModelAttr:    e.model.Attr,
				LookForward:  e.lookForward,
				LookBackward: e.model.LookBackward,
			}).
			Post(e.model.SourceImplURL + consts.TRAIN_SUFFIX)
		if err != nil {
			asyncErrors <- err
			log.Logger.Error(err, "")
			return
		}
		if response.StatusCode() != http.StatusOK {
			asyncErrors <- errors.New(string(response.Body()))
			return
		}
		finishChan <- struct{}{}
	}(finishChan, asyncErrors)
	return finishChan, asyncErrors
}

func (e *ExternalPredictorImpl) Loss() (float64, error) {
	lookForward := e.lookForward
	lookBackward := e.model.LookBackward
	//该数据前半段作为预测的输入，后半段作为验证的真实值
	metricData := e.facade.WaitToGetMetric(e.wmk.ToNoModelKey(), lookForward+lookBackward)
	client := resty.New()
	response, err := client.R().
		SetBody(message.PredictRequest{
			Metrics:      metricData[:lookBackward],
			Key:          e.wmk,
			ModelAttr:    e.model.Attr,
			LookForward:  e.lookForward,
			LookBackward: e.model.LookBackward,
		}).
		Post(e.model.SourceImplURL + consts.PREDICT_SUFFIX)
	if err != nil {
		log.Logger.Error(err, "")
		return -1, err
	}
	if response.StatusCode() != http.StatusOK {
		return -1, errors.New(string(response.Body()))
	}
	result := message.PredictResult{}
	err = json.Unmarshal(response.Body(), &result)
	if err != nil {
		log.Logger.Error(err, "")
		return -1, err
	}
	if e.model.Debug {
		_ = utils.PlotTwoLines(metricData[len(metricData)-lookForward:], result.PredictMetric, "loss")
	}
	return utils.GetMSELoss(metricData[len(metricData)-lookForward:], result.PredictMetric), nil
}

func NewExternalPredictor(param message.Param) ExternalPredictor {
	return &ExternalPredictorImpl{
		wmk:         param.WithModelKey,
		facade:      param.CollectorFacade,
		lookForward: param.LookForward,
		model:       param.Model,
	}
}
