package controller

import (
	"context"
	"errors"
	elasticv1 "github.com/LL-res/CRM/api/v1"
	"github.com/LL-res/CRM/clients/k8s"
	"github.com/LL-res/CRM/collector"
	"github.com/LL-res/CRM/common/consts"
	"github.com/LL-res/CRM/common/key"
	"github.com/LL-res/CRM/common/log"
	"github.com/LL-res/CRM/domain/BO"
	"github.com/LL-res/CRM/knowledge"
	"github.com/LL-res/CRM/predictor"
	"github.com/LL-res/CRM/predictor/message"
	"github.com/LL-res/CRM/scheduler"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	"time"
)

type Controller struct {
	instance *elasticv1.CRM
	//predictors map[*basetype.Metric][]predictor.Predictor
	*CRMReconciler
}

func NewController(instance *elasticv1.CRM, reconciler *CRMReconciler) *Controller {
	return &Controller{
		instance:      instance,
		CRMReconciler: reconciler,
	}
}

func (hdlr *Controller) Handle(ctx context.Context) error {
	// 是由 status的更新导致
	if hdlr.instance.Status.Generation == hdlr.instance.Generation {
		return nil
	}
	if err := k8s.NewClient(); err != nil {
		return err
	}
	// 防止过多层if嵌套
	var err error
	// create instance
	if hdlr.instance.Status.Generation == 0 {
		log.Logger.Info("creating aom instance", "namespace", hdlr.instance.Namespace, "name", hdlr.instance.Name)
		err = hdlr.handleCreate(ctx)
	}
	// update instance
	if hdlr.instance.Status.Generation != 0 &&
		hdlr.instance.Generation > hdlr.instance.Status.Generation {
		log.Logger.Info("updating aom instance", "namespace", hdlr.instance.Namespace, "name", hdlr.instance.Name)
		err = hdlr.handleUpdate(ctx)
	}
	if err != nil {
		return err
	}

	hdlr.instance.Status.Generation = hdlr.instance.Generation

	if err := hdlr.Status().Update(ctx, hdlr.instance); err != nil {
		log.Logger.Error(err, "update status failed")
		return err
	}
	return nil

}

func (hdlr *Controller) handleUpdate(ctx context.Context) error {
	if err := hdlr.handleMetrics(ctx); err != nil {
		return err
	}
	changes, err := hdlr.handleModels(ctx)
	if err != nil {
		return err
	}
	if err := hdlr.handleCollector(ctx); err != nil {
		return err
	}
	if err := hdlr.handlePredictor(ctx, changes); err != nil {
		return err
	}
	return nil
}

func (hdlr *Controller) handleCreate(ctx context.Context) error {
	//promcOnce.Do(func() {
	//	log.Logger.Info("init collector")
	//	pCollector = prometheus_collector.New(ctx, hdlr.instance.Spec.Collector.BaseOnHistory, hdlr.instance.Spec.Collector.ScrapeInterval)
	//})
	//err := pCollector.SetServerAddress(hdlr.instance.Spec.Collector.Address)
	collectrFacade, err := collector.GetOrNewFacade(time.Duration(hdlr.instance.Spec.Collector.ScrapeInterval*hdlr.instance.Spec.IntervalDuration)*time.Second,
		hdlr.instance.Spec.Collector.BaseOnHistory,
		hdlr.instance.Spec.Collector.Address,
		consts.PROMETHEUS,
		hdlr.instance.Spec.Collector.MaxCap,
	)
	if err != nil {
		log.Logger.Error(err, "fail to set collector server address")
		return err
	}
	localKnowledge := knowledge.GetLocalKnowledge(types.NamespacedName{
		Namespace: ctx.Value(consts.NAMESPACE).(string),
		Name:      ctx.Value(consts.NAME).(string),
	})
	localKnowledge.CollectorFacade = collectrFacade
	localKnowledge.Scaler = localKnowledge.Scaler.New(ctx.Value(consts.NAMESPACE).(string), hdlr.instance.Spec.ScaleTargetRef, hdlr.instance.Spec.MaxReplicas, hdlr.instance.Spec.MinReplicas)
	log.Logger.Info("init scaler", "scaler", localKnowledge.Scaler)
	if err := hdlr.handleMetrics(ctx); err != nil {
		return err
	}

	changes, err := hdlr.handleModels(ctx)
	if err != nil {
		return err
	}

	if err := hdlr.handleCollector(ctx); err != nil {
		return err
	}

	if err := hdlr.handlePredictor(ctx, changes); err != nil {
		return err
	}
	schdlr := scheduler.GetOrNew(types.NamespacedName{
		Namespace: ctx.Value(consts.NAMESPACE).(string),
		Name:      ctx.Value(consts.NAME).(string),
	}, time.Second*time.Duration(hdlr.instance.Spec.IntervalDuration), hdlr.instance.Spec.Models.LookForward)
	log.Logger.Info("start scheduler", "scheduler", schdlr)
	go schdlr.Start(ctx)

	return nil
}

func (hdlr *Controller) handleDelete(ctx context.Context) error {
	return nil
}

func (hdlr *Controller) handleCollector(ctx context.Context) error {
	// 此操作为幂等操作
	// 其中的元素是格式化之后的metric，格式为: name$unit$query
	toDelete := make([]key.NoModelKey, 0)
	toAdd := make([]BO.Metric, 0)
	localKnowledge := knowledge.GetLocalKnowledge(types.NamespacedName{
		Namespace: ctx.Value(consts.NAMESPACE).(string),
		Name:      ctx.Value(consts.NAME).(string),
	})
	// spec中存在，但map中不存在，进行更新
	collectorSetForNow := localKnowledge.CollectorFacade.GetCollectorKeySet()
	for _, metric := range hdlr.instance.Spec.Metrics {

		if _, ok := collectorSetForNow[metric.NoModelKey()]; !ok {
			toAdd = append(toAdd, metric)
		}
	}
	// map 中存在但 spec中不存在，进行删除
	for k := range collectorSetForNow {
		exist := false
		for _, metric := range hdlr.instance.Spec.Metrics {
			if metric.NoModelKey() == k {
				exist = true
				break
			}
		}
		if !exist {
			toDelete = append(toDelete, k)
		}
	}
	for _, v := range toDelete {
		log.Logger.Info("delete metric worker", "metric", v)
		// 对collecter worker进行退出控制
		//close(localKnowledge.CollectorMap[v])
		//localKnowledge.CollectorWorkerMap.Delete(v)
		localKnowledge.CollectorFacade.DeleteCollector(v)
	}
	for _, m := range toAdd {
		localKnowledge.CollectorFacade.CreateCollector(m.NoModelKey())
		//pCollector.AddCustomMetrics(m)
		//worker, err := pCollector.CreateWorker(m)
		//if err != nil {
		//	log.Logger.Error(err, "fail to create metric collector worker")
		//	return err
		//}
		//log.Logger.Info("create metric worker", "metric key", m.NoModelKey())
		//localKnowledge.CollectorWorkerMap.Store(m.NoModelKey(), worker)
		//stopC := make(chan struct{})
		//localKnowledge.CollectorMap[m.NoModelKey()] = stopC
		//go worker.Collect(stopC)
		//go StartWorker(ctx, worker, hdlr.instance, stopC)
	}
	// 更新status
	//hdlr.instance.Status.StatusCollectors = make([]automationv1.StatusCollector, 0, len(hdlr.instance.Spec.Metrics))
	//for _, metric := range hdlr.instance.Spec.Metrics {
	//	// 此处仅作describe时显示
	//	hdlr.instance.Status.StatusCollectors = append(hdlr.instance.Status.StatusCollectors, automationv1.StatusCollector{
	//		Name:       metric.Name,
	//		Unit:       metric.Unit,
	//		Expression: metric.Query,
	//	})
	//}
	return nil
}

type mdlMtrc struct {
	model  BO.Model
	metric BO.Metric
}

func (hdlr *Controller) handlePredictor(ctx context.Context, changeMap map[key.WithModelKey]BO.Model) error {
	localKnowledge := knowledge.GetLocalKnowledge(types.NamespacedName{
		Namespace: ctx.Value(consts.NAMESPACE).(string),
		Name:      ctx.Value(consts.NAME).(string),
	})
	// 扫一遍spec 查看现在所需的

	// sepc 中存在，map中不存在
	toAdd := make([]mdlMtrc, 0)
	for key, models := range hdlr.instance.Spec.Models.ModelsForMetric {
		metric, ok := hdlr.instance.Spec.Metrics[key]
		if !ok {
			// TODO validation
			log.Logger.Error(errors.New("not found metric"), "validate failed")
			return errors.New("not found metric")
		}
		for _, model := range models {
			if _, err := localKnowledge.PredictorMap.Load(metric.WithModelKey(model.Type)); err != nil {
				toAdd = append(toAdd, mdlMtrc{
					model:  model,
					metric: metric,
				})
			}
		}
	}

	// map 中存在，spec中不存在
	toDelete := make([]key.WithModelKey, 0)
	// 先将spec中的key都放入tempMap中，再进行比较以降低复杂度
	tempMap := make(map[key.WithModelKey]struct{})
	for key, models := range hdlr.instance.Spec.Models.ModelsForMetric {
		metric, ok := hdlr.instance.Spec.Metrics[key]
		if !ok {
			// TODO validation
			log.Logger.Error(errors.New("not found metric"), "validate failed")
			return errors.New("not found metric")
		}
		for _, model := range models {
			tempMap[metric.WithModelKey(model.Type)] = struct{}{}
		}
	}

	for k := range localKnowledge.PredictorMap.Data {
		if _, ok := tempMap[k]; !ok {
			toDelete = append(toDelete, k)
		}
	}
	for _, wmk := range toDelete {
		log.Logger.Info("delete predictor", "predictor", wmk)
		localKnowledge.PredictorMap.Delete(wmk)
		//nmk := utils.GetNoModelKey(wmk)
		//找到metric对应的那一组predictor
		//metric, err := hdlr.instance.MetricMap.Get(nmk)
		//if err != nil {
		//	log.Logger.Error(err, "a must behaviour failed,predictor can not find the corresponding metric")
		//	return err
		//}
		//for i, pred := range hdlr.predictors[metric] {
		//	if pred.Key() == wmk {
		//		hdlr.predictors[metric] = append(hdlr.predictors[metric][:i], hdlr.predictors[metric][i+1:]...)
		//	}
		//}
	}
	for _, metricModelPair := range toAdd {
		log.Logger.Info("create predictor", "predictor", metricModelPair.metric.WithModelKey(metricModelPair.model.Type))
		WithModelKey := metricModelPair.metric.WithModelKey(metricModelPair.model.Type)
		pm := message.Param{
			LookForward:     hdlr.instance.Spec.Models.LookForward,
			WithModelKey:    WithModelKey,
			CollectorFacade: localKnowledge.CollectorFacade,
			Model:           metricModelPair.model,
		}
		pred, err := predictor.NewPredictor(pm)
		if err != nil {
			log.Logger.Error(err, "new predictor failed", "metricModelPair", pm)
			return err
		}
		localKnowledge.PredictorMap.Store(metricModelPair.metric.WithModelKey(metricModelPair.model.Type), pred)
		//metric, err := hdlr.instance.MetricMap.Get(metricModelPair.NoModelKey())
		//if err != nil {
		//	log.Logger.Error(err, "a must behaviour failed,predictor can not find the corresponding metric")
		//	return err
		//}
		//hdlr.predictors[metric] = append(hdlr.predictors[metric], pred)
	}
	for wmk, model := range changeMap {
		pred, err := predictor.NewPredictor(message.Param{
			LookForward:     hdlr.instance.Spec.Models.LookForward,
			WithModelKey:    wmk,
			CollectorFacade: localKnowledge.CollectorFacade,
			Model:           model,
		})
		if err != nil {
			log.Logger.Error(err, "new predictor failed")
			return err
		}
		localKnowledge.PredictorMap.Store(wmk, pred)
	}
	// TODO STATUS
	return nil
}

func (hdlr *Controller) handleMetrics(ctx context.Context) error {
	localKnowledge := knowledge.GetLocalKnowledge(types.NamespacedName{
		Namespace: ctx.Value(consts.NAMESPACE).(string),
		Name:      ctx.Value(consts.NAME).(string),
	})
	// Add
	for _, metric := range hdlr.instance.Spec.Metrics {
		if _, err := localKnowledge.MetricMap.Load(metric.NoModelKey()); err != nil {
			log.Logger.Info("store metric", "metric", metric.NoModelKey())
			localKnowledge.MetricMap.Store(metric.NoModelKey(), &metric)
		}
	}
	// Delete
	tempMap := make(map[key.NoModelKey]struct{})
	for _, metric := range hdlr.instance.Spec.Metrics {
		tempMap[metric.NoModelKey()] = struct{}{}
	}
	for nmk := range localKnowledge.MetricMap.Data {
		if _, ok := tempMap[nmk]; !ok {
			log.Logger.Info("delete metric", "metric", nmk)
			localKnowledge.MetricMap.Delete(nmk)
		}
	}
	//change
	for _, metric := range hdlr.instance.Spec.Metrics {
		old, err := localKnowledge.MetricMap.Load(metric.NoModelKey())
		if err != nil {
			log.Logger.Error(err, "a must behaviour failed")
			return err
		}
		if reflect.DeepEqual(*old, metric) {
			return nil
		}
		log.Logger.Info("update metric", "metric", metric.NoModelKey())
		localKnowledge.MetricMap.Store(metric.NoModelKey(), &metric)
	}
	return nil
}

func (hdlr *Controller) handleModels(ctx context.Context) (map[key.WithModelKey]BO.Model, error) {
	localKnowledge := knowledge.GetLocalKnowledge(types.NamespacedName{
		Namespace: ctx.Value(consts.NAMESPACE).(string),
		Name:      ctx.Value(consts.NAME).(string),
	})
	//delete
	tempMap := make(map[key.WithModelKey]struct{})
	//Add or change
	changeMap := make(map[key.WithModelKey]BO.Model)
	for specKey, models := range hdlr.instance.Spec.Models.ModelsForMetric {
		metric, ok := hdlr.instance.Spec.Metrics[specKey]
		if !ok {
			return nil, errors.New("orphan model")
		}
		for _, model := range models {
			wmk := metric.WithModelKey(model.Type)
			tempMap[wmk] = struct{}{}
			old, err := localKnowledge.ModelMap.Load(wmk)
			if err != nil {
				log.Logger.Info("store model", "model", wmk)
				modelToStore := model
				localKnowledge.ModelMap.Store(wmk, &modelToStore)
				continue
			}
			if reflect.DeepEqual(*old, model) {
				continue
			}
			log.Logger.Info("update model", "model", wmk)
			localKnowledge.ModelMap.Store(wmk, &model)
			changeMap[wmk] = model
		}
	}
	//delete
	for wmk := range localKnowledge.ModelMap.Data {
		if _, ok := tempMap[wmk]; !ok {
			log.Logger.Info("delete model", "model", wmk)
			localKnowledge.ModelMap.Delete(wmk)
		}
	}
	return changeMap, nil
}
