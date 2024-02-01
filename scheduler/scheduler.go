package scheduler

import (
	"context"
	"fmt"
	"github.com/LL-res/CRM/common/errs"
	"github.com/LL-res/CRM/common/key"
	"github.com/LL-res/CRM/common/log"
	"github.com/LL-res/CRM/knowledge"
	"github.com/LL-res/CRM/scaler"
	"strconv"
	"time"

	"k8s.io/apimachinery/pkg/types"
)

type Scheduler struct {
	Name             types.NamespacedName
	Interval         time.Duration
	lookForward      int
	sortedPredictors MetricToPredictors
}
type Conf struct {
	needTrain       bool
	trainInterval   time.Duration
	predictInterval time.Duration
}

var schedulers map[types.NamespacedName]*Scheduler

func GetOrNew(name types.NamespacedName, interval time.Duration, lookForward int) *Scheduler {
	if nil == schedulers {
		schedulers = make(map[types.NamespacedName]*Scheduler)
	}
	if nil == schedulers[name] {
		schedulers[name] = New(name, interval, lookForward)
	}
	return schedulers[name]
}
func New(name types.NamespacedName, interval time.Duration, lookForward int) *Scheduler {
	return &Scheduler{
		Name: name,
		// the interval AOM call all the models
		Interval:         interval,
		sortedPredictors: NewMetricToPredictors(),
		lookForward:      lookForward,
	}
}

type ResPair struct {
	modelReplica []int32
	withModelKey string
}

func (s *Scheduler) DeepCopyInto(out *Scheduler) {

}

func (s *Scheduler) Start(ctx context.Context) {
	//hide contains all the metrics,models and predictors that the instance owns
	localKnowledge := knowledge.GetLocalKnowledge(s.Name)
	//use a block procession to put math predictors to the heap
	startTime := time.Now()
	s.LoadAvailablePredictors(localKnowledge)
	endTime := time.Now()
	consumingDur := endTime.Sub(startTime)
	log.Logger.Info("All untrainable model loaded", "time consuming", fmt.Sprintf("%dm %ds", int(consumingDur.Seconds())/60, int(consumingDur.Seconds())%60))
	//start goroutine to train all the predictors background
	go s.UpdatePredictors(ctx)
	//start a reactive scaler
	//predict every look forward interval,and set the schedule for next lookForward interval
	ticker := time.NewTicker(time.Duration(s.lookForward) * s.Interval)
	defer ticker.Stop()
	for ; true; <-ticker.C {
		log.Logger.Info("starting to manage next period", "ts", time.Now().Format("15:04:05"))
		//the result of one period replicas
		//using all the metrics to get this
		onePeriodReplicas := make([]int32, s.lookForward)
		//some metrics may failed to calculate the replica
		//total weight only contains the weight whose metric has done the process successfully
		var totalWeight int32
		//it's time to predict
		//iterate the list of heaps
		for nmk, predictors := range s.sortedPredictors {
			//use the min loss predictor to predict
			wmk := predictors.Peek()
			//the metric has no availabe predictor
			if wmk.IsEmpty() {
				continue
			}
			pred, err := localKnowledge.PredictorMap.Load(wmk)
			if err != nil {
				log.Logger.Error(err, "failed to find the predictors")
				continue
			}
			log.Logger.Info("current predictor", "key", wmk)
			pResult, err := pred.Predict(ctx)
			if err == errs.NO_SUFFICENT_DATA || err == errs.UNREADY_TO_PREDICT {
				log.Logger.Info("the predictor needs more metrics to be funtional", "predictor", wmk)
				continue
			}
			if err != nil {
				log.Logger.Error(err, "failed to predict", "key", wmk)
				continue
			}
			//use the metric to find the target
			//and then ,calculate the replicas
			metric, err := localKnowledge.MetricMap.Load(nmk)
			if err != nil {
				log.Logger.Error(err, "")
				continue
			}
			targetVal, err := strconv.ParseFloat(metric.Target, 64)
			if err != nil {
				log.Logger.Error(err, "strconv failed")
				return
			}
			//get the replicas that one model predicted
			predictReplicas, err := localKnowledge.Scaler.GetModelReplica(pResult.PredictMetric, pResult.StartMetric, scaler.UnderThresholdPerPod, targetVal)
			if err != nil {
				log.Logger.Error(err, "")
				continue
			}
			//multiply the weight of the metric and divide 100 when all the metric have been iterated
			for i := range onePeriodReplicas {
				onePeriodReplicas[i] += predictReplicas[i] * metric.Weight
			}
			//the metric has finished the process and calculate the replicas
			totalWeight += metric.Weight
		}
		//there is no metric configured,continue to wait
		if totalWeight == 0 {
			continue
		}
		//all the models have calculated the replicas
		for i := range onePeriodReplicas {
			onePeriodReplicas[i] = onePeriodReplicas[i] / totalWeight
		}
		log.Logger.Info("one period replicas", "replicas", onePeriodReplicas)
		//use the scaler to arrange the replica for the managed deployment
		go localKnowledge.Scaler.ManageOnePeriod(onePeriodReplicas, s.Interval)
	}
}

func (s *Scheduler) UpdatePredictors(ctx context.Context) {
	//use the interval to tick
	ticker := time.NewTicker(time.Duration(3*s.lookForward) * s.Interval)
	localKnowledge := knowledge.GetLocalKnowledge(s.Name)
	trainingSet := make(map[key.WithModelKey]struct{})
	defer ticker.Stop()
	for ; ; <-ticker.C {
		for wmk, pred := range localKnowledge.PredictorMap.Data {
			model, err := localKnowledge.ModelMap.Load(wmk)
			if err != nil {
				log.Logger.Error(err, "no model found")
				continue
			}
			//the model is not trainable
			//e.g. math models
			if !model.NeedTrain {
				predictorHeap := s.sortedPredictors.GetHeap(wmk.ToNoModelKey())
				//if the predictor exists in the heap,use the min loss model
				oldPredictor := predictorHeap.find(wmk)
				predLoss, _ := pred.Loss()
				if oldPredictor != nil {
					predictorHeap.update(oldPredictor, wmk, predLoss)
					log.Logger.Info("update a predictor", "key", wmk, "loss", predLoss)
					continue
				}
				//if the predictor does not exist in the heap,push the item
				newItem := &Item{
					value:    wmk,
					priority: predLoss,
					index:    predictorHeap.Len(),
				}
				predictorHeap.Push(newItem)
				predictorHeap.update(newItem, wmk, predLoss)
				continue
			}
			lastTime, err := localKnowledge.TrainHistory.Load(wmk)
			//need train or not
			//if need to train the trainable model
			needTrain := false
			//the model has never been trained before,then train
			//the model has been trained ,one update interval has passed,the it should retrain
			//if the model is training ,then skip
			if _, ok := trainingSet[wmk]; ok {
				continue
			}
			if err != nil || lastTime.Add(s.Interval*time.Duration(model.UpdateInterval)).After(time.Now()) {
				needTrain = true
			}
			if !needTrain {
				//update the model loss in heap
				newLoss, _ := pred.Loss()
				predictorHeap := s.sortedPredictors.GetHeap(wmk.ToNoModelKey())
				oldPredictor := predictorHeap.find(wmk)
				predictorHeap.update(oldPredictor, wmk, newLoss)
				log.Logger.Info("update a predictor", "key", wmk, "loss", newLoss)
				continue
			}
			//mark the model as training
			trainingSet[wmk] = struct{}{}
			// train use asynchronous,so it won`t block the process for too long
			finishedChan, errChan := pred.Train(ctx)
			go s.afterTrain(trainingSet, finishedChan, errChan, wmk)
		}
	}
}

func (s *Scheduler) LoadAvailablePredictors(localKnowledge *knowledge.LocalKnowledge) {
	for wmk, pred := range localKnowledge.PredictorMap.Data {
		model, err := localKnowledge.ModelMap.Load(wmk)
		if err != nil {
			log.Logger.Error(err, "")
			continue
		}
		//load all the models which don't need to train
		if !model.NeedTrain || model.PreTrained {
			loss, _ := pred.Loss()
			predictorHeap := s.sortedPredictors.GetHeap(wmk.ToNoModelKey())
			newItem := &Item{
				value:    wmk,
				priority: loss,
				index:    predictorHeap.Len(),
			}
			predictorHeap.Push(newItem)
			predictorHeap.update(newItem, wmk, loss)
			log.Logger.Info("activate a predictor", "key", wmk, "loss", loss)
		}
		if model.PreTrained {
			localKnowledge.TrainHistory.Store(wmk, time.Now())
		}
	}
}

func (s *Scheduler) afterTrain(trainingSet map[key.WithModelKey]struct{}, finishedChan chan struct{}, errorsChan chan error, wmk key.WithModelKey) {
	select {
	case <-finishedChan:
		// now the training finished
		localKnowledge := knowledge.GetLocalKnowledge(s.Name)
		localKnowledge.TrainHistory.Store(wmk, time.Now())
		predictorHeap := s.sortedPredictors.GetHeap(wmk.ToNoModelKey())
		//if the predictor exists in the heap,use the min loss model
		oldPredictor := predictorHeap.find(wmk)
		pred, _ := localKnowledge.PredictorMap.Load(wmk)
		predLoss, _ := pred.Loss()
		if oldPredictor != nil {
			predictorHeap.update(oldPredictor, wmk, predLoss)
			log.Logger.Info("update a predictor", "key", wmk, "loss", predLoss)
			delete(trainingSet, wmk)
			return
		}
		//if the predictor does not exist in the heap,push the item
		newItem := &Item{
			value:    wmk,
			priority: predLoss,
			index:    predictorHeap.Len(),
		}
		predictorHeap.Push(newItem)
		predictorHeap.update(newItem, wmk, predLoss)
		log.Logger.Info("activate a predictor", "key", wmk, "loss", predLoss)
		delete(trainingSet, wmk)
	case err := <-errorsChan:
		delete(trainingSet, wmk)
		if err == errs.NO_SUFFICENT_DATA {
			if err.(*errs.NoSufficentErr).GetCap()%50 == 0 {
				log.Logger.Info("data preparing", "key", wmk, "data size", err.Error())
				return
			}
		}
		log.Logger.Error(err, "train model failed", "key", wmk)
	}
}
