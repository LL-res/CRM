package holt_winter

import (
	"context"
	"errors"
	"github.com/LL-res/CRM/collector"
	"github.com/LL-res/CRM/common/consts"
	"github.com/LL-res/CRM/common/errs"
	"github.com/LL-res/CRM/common/key"
	"github.com/LL-res/CRM/common/log"
	"github.com/LL-res/CRM/predictor/message"
	"github.com/LL-res/CRM/utils"
	"strconv"
	"time"
)

type HoltWinter struct {
	debug           bool
	slen            int
	lookForward     int
	lookBackward    int
	alpha           float64
	beta            float64
	gamma           float64
	withModelKey    key.WithModelKey
	collectorFacade collector.CollectorFacade
	loss            float64
}

type Param struct {
	Slen         string `json:"slen,omitempty"`
	LookForward  string `json:"look_forward,omitempty"`
	LookBackward string `json:"look_backward,omitempty"`
	Alpha        string `json:"alpha,omitempty"`
	Beta         string `json:"beta,omitempty"`
	Gamma        string `json:"gamma,omitempty"`
	Debug        string `json:"debug,omitempty"`
}

func (p *HoltWinter) Predict(ctx context.Context) (message.PredictResult, error) {
	if p.collectorFacade.GetCapFromCollector(p.withModelKey.ToNoModelKey()) < p.lookBackward {
		return message.PredictResult{}, errs.NO_SUFFICENT_DATA
	}
	metrics := p.collectorFacade.GetMetricFromCollector(p.withModelKey.ToNoModelKey(), p.lookBackward)
	metrics = metrics[len(metrics)-p.lookBackward:]

	predMetrics := p.tripleExponentialSmoothing(metrics)
	if p.debug {
		//log.Logger.Info("predict metrics", "metrics", predMetrics)
		if err := utils.PlotLine(metrics, predMetrics, "holt_winter"); err != nil {
			log.Logger.Error(err, "debug plot failed")
		}
	}
	res := message.PredictResult{
		StartMetric:   metrics[len(metrics)-1],
		Loss:          -1,
		PredictMetric: predMetrics,
	}
	return res, nil
}

func (p *HoltWinter) GetType() string {
	return consts.HOLT_WINTER
}

func (p *HoltWinter) Train(ctx context.Context) (chan struct{}, chan error) {
	return nil, nil
}

func (p *HoltWinter) Key() key.WithModelKey {
	return p.withModelKey
}

func (p *HoltWinter) initialTrend(series []float64) float64 {
	sum := 0.0
	for i := 0; i < p.slen; i++ {
		sum += (series[i+p.slen] - series[i]) / float64(p.slen)
	}
	return sum / float64(p.slen)
}

func (p *HoltWinter) initialSeasonalComponents(series []float64) map[int]float64 {
	seasonals := make(map[int]float64)
	seasonAverages := make([]float64, 0)
	nSeasons := len(series) / p.slen

	// Compute season averages
	for j := 0; j < nSeasons; j++ {
		sum := 0.0
		for _, value := range series[p.slen*j : p.slen*j+p.slen] {
			sum += value
		}
		seasonAverages = append(seasonAverages, sum/float64(p.slen))
	}

	// Compute initial values
	for i := 0; i < p.slen; i++ {
		sumOfValsOverAvg := 0.0
		for j := 0; j < nSeasons; j++ {
			sumOfValsOverAvg += series[p.slen*j+i] - seasonAverages[j]
		}
		seasonals[i] = sumOfValsOverAvg / float64(nSeasons)
	}

	return seasonals
}
func (p *HoltWinter) Loss() (float64, error) {
	//wait until there is sufficient data to do a prediction and validate
	for p.collectorFacade.GetCapFromCollector(p.withModelKey.ToNoModelKey()) < p.lookForward+p.lookBackward {
		time.Sleep(30 * time.Second)
	}
	//will not scale.only do a prediction and find out the loss of the predictor
	predictValidateData := p.collectorFacade.GetMetricFromCollector(p.withModelKey.ToNoModelKey(), p.lookForward+p.lookBackward)
	useToPredict := predictValidateData[:p.lookBackward]
	useToValidate := predictValidateData[len(predictValidateData)-p.lookForward:]
	predMetrics := p.tripleExponentialSmoothing(useToPredict)
	if p.debug {
		_ = utils.PlotTwoLines(useToValidate, predMetrics, "loss")
	}
	return utils.GetMSELoss(predMetrics, useToValidate), nil
}
func Normalize(nums []float64) []float64 {
	sum := 0.0
	for _, v := range nums {
		sum += v
	}
	res := make([]float64, 0)
	for _, v := range nums {
		res = append(res, v/sum)
	}
	return res
}
func (p *HoltWinter) tripleExponentialSmoothing(series []float64) []float64 {
	result := make([]float64, 0)
	seasonals := p.initialSeasonalComponents(series)
	smooth := series[0]
	trend := p.initialTrend(series)

	for i := 0; i < len(series)+p.lookForward; i++ {
		if i == 0 {
			result = append(result, series[0])
			continue
		}
		if i >= len(series) {
			m := i - len(series) + 1
			result = append(result, (smooth+float64(m)*trend)+seasonals[i%p.slen])
		} else {
			val := series[i]
			lastSmooth := smooth
			smooth = p.alpha*(val-seasonals[i%p.slen]) + (1-p.alpha)*(smooth+trend)
			trend = p.beta*(smooth-lastSmooth) + (1-p.beta)*trend
			seasonals[i%p.slen] = p.gamma*(val-smooth) + (1-p.gamma)*seasonals[i%p.slen]
			result = append(result, smooth+trend+seasonals[i%p.slen])
		}
	}

	return result[len(result)-p.lookForward:]
}

func New(param message.Param) (*HoltWinter, error) {
	slenStr, ok := param.Model.Attr["slen"]
	if !ok {
		return nil, errors.New("parameter missing")
	}
	alphaStr, ok := param.Model.Attr["alpha"]
	if !ok {
		return nil, errors.New("parameter missing")
	}
	betaStr, ok := param.Model.Attr["beta"]
	if !ok {
		return nil, errors.New("parameter missing")
	}
	gammaStr, ok := param.Model.Attr["gamma"]
	if !ok {
		return nil, errors.New("parameter missing")
	}
	slen, err := strconv.Atoi(slenStr)
	if err != nil {
		return nil, err
	}
	alpha, err := strconv.ParseFloat(alphaStr, 64)
	if err != nil {
		return nil, err
	}
	beta, err := strconv.ParseFloat(betaStr, 64)
	if err != nil {
		return nil, err
	}
	gamma, err := strconv.ParseFloat(gammaStr, 64)
	if err != nil {
		return nil, err
	}
	return &HoltWinter{
		slen:            slen,
		lookForward:     param.LookForward,
		lookBackward:    param.Model.LookBackward,
		alpha:           alpha,
		beta:            beta,
		gamma:           gamma,
		withModelKey:    param.WithModelKey,
		collectorFacade: param.CollectorFacade,
		debug:           param.Model.Debug,
		loss:            -1,
	}, nil

}
