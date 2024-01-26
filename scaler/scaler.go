package scaler

import (
	"fmt"
	"github.com/LL-res/AOM/utils"
	"time"

	"github.com/LL-res/AOM/clients/k8s"
	"github.com/LL-res/AOM/log"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
)

type Scaler struct {
	MaxReplica     int32  `json:"maxReplica"`
	MinReplica     int32  `json:"minReplica"`
	Namespace      string `json:"namespace"`
	ScaleTargetRef autoscalingv2.CrossVersionObjectReference
	recvChan       chan []float64
	//plant a seed that record the first time the scaler wanted to scale down
	ScaleDownSeed *time.Time
}

func (s *Scaler) RecvChan() chan []float64 {
	if s.recvChan == nil {
		s.recvChan = make(chan []float64)
	}
	return s.recvChan
}
func (s *Scaler) New(namespace string, scaleTargetRef autoscalingv2.CrossVersionObjectReference, maxReplica, minReplica int32) *Scaler {
	return &Scaler{MinReplica: minReplica, MaxReplica: maxReplica, Namespace: namespace, ScaleTargetRef: scaleTargetRef}
}
func New(namespace string, scaleTargetRef autoscalingv2.CrossVersionObjectReference, maxReplica, minReplica int32) *Scaler {

	return &Scaler{MinReplica: minReplica, MaxReplica: maxReplica, Namespace: namespace, ScaleTargetRef: scaleTargetRef}

}

// 每个model对应一个
func (s *Scaler) GetModelReplica(predictMetrics []float64, startMetric float64, strategy BaseStrategy, targetMetric float64) ([]int32, error) {
	startReplica, err := k8s.GlobalClient.GetReplica(s.Namespace, s.ScaleTargetRef)
	if err != nil {
		return nil, err
	}
	// 获取预测的指标数据
	return strategy(targetMetric, startMetric, startReplica, predictMetrics), nil

}

// 获取每个metric对应的预测样本数
func (s *Scaler) GetMetricReplica(modelReplica [][]int32, strategy ModelStrategy) []int32 {
	return strategy(modelReplica)
}

// 获取到了被检测对象之后时间端的样本数
// 之后的操作应该是从这个切片中进行选取，选取一个或是多个合适的值，作为在当前时刻要进行的扩缩容副本数
func (s *Scaler) GetObjReplica(metricReplica [][]int32, strategy MetricStrategy) []int32 {
	return strategy(metricReplica)
}

func (s *Scaler) GetScaleReplica(objReplicaSet []int32, strategy ObjStrategy) int32 {
	return strategy(objReplicaSet)
}
func (s *Scaler) UpTo(replica int32) error {
	curReplica, err := k8s.GlobalClient.GetReplica(s.Namespace, s.ScaleTargetRef)
	if err != nil {
		return err
	}
	if curReplica >= replica {
		log.Logger.Info("do not scale", "scale target", s.ScaleTargetRef, "current replica", fmt.Sprint(curReplica), "target replica", fmt.Sprint(replica))
		return ErrTargetSmallerThanCurrent
	}
	if replica > s.MaxReplica {
		log.Logger.Info("scale to max replica", "scale target", s.ScaleTargetRef, "max replica", fmt.Sprint(s.MaxReplica), "target replica", fmt.Sprint(replica))
		replica = s.MaxReplica
	}
	err = k8s.GlobalClient.SetReplica(s.Namespace, s.ScaleTargetRef, replica)
	if err != nil {
		return err
	}
	return nil

}
func (s *Scaler) CheckSeed(dur int) bool {
	if s.ScaleDownSeed == nil {
		x := time.Now()
		s.ScaleDownSeed = &x
		return false
	}
	if (*s.ScaleDownSeed).Add(time.Duration(dur) * time.Second).Before(time.Now()) {
		return true
	}
	return false
}
func (s *Scaler) DownWithStep(step int32) error {
	curReplica, err := k8s.GlobalClient.GetReplica(s.Namespace, s.ScaleTargetRef)
	replica := curReplica - step
	if err != nil {
		return err
	}
	if curReplica <= replica {
		return ErrTargetBiggerThanCurrent
	}
	if replica < s.MinReplica {
		log.Logger.Info("scale to min replica", "scale target", s.ScaleTargetRef, "min replica", fmt.Sprint(s.MinReplica), "target replica", fmt.Sprint(replica))
		replica = s.MinReplica
	}
	err = k8s.GlobalClient.SetReplica(s.Namespace, s.ScaleTargetRef, replica)
	if err != nil {
		return err
	}
	s.ScaleDownSeed = nil
	return nil
}
func (s *Scaler) ManageOnePeriod(replicas []int32, interval time.Duration) {
	infector := NewInfector("SlopeWindow")
	logReplica, sleepTime := infector.Aggregate(replicas, interval, 0)
	logReplicaWithTime := make([]string, 0)
	//log.Logger.Info("scaler is scaling", "replicas", logReplica, "interval", fmt.Sprintf("%fs", sleepTime.Seconds()))
	for _, v := range logReplica {
		if v > s.MaxReplica {
			v = s.MaxReplica
		}
		if v < s.MinReplica {
			v = s.MinReplica
		}
		//log.Logger.Info("scale start time point", "ts", time.Now().Format("15:04:05"))
		err := k8s.GlobalClient.SetReplica(s.Namespace, s.ScaleTargetRef, v)
		//log.Logger.Info("scale finish time point", "ts", time.Now().Format("15:04:05"))
		if err != nil {
			log.Logger.Error(err, "period management failed")
			return
		}
		logReplicaWithTime = append(logReplicaWithTime, fmt.Sprintf("ts : %s,replica : %d", time.Now().Format("15:04:05"), v))
		time.Sleep(sleepTime)
	}
	log.Logger.Info("one period managed", "replicas", logReplicaWithTime)

}
func infectWithWindowMax(replicas []int32, interval time.Duration, windowSize int) ([]int32, time.Duration) {
	if windowSize == 0 || windowSize == 1 {
		return replicas, interval
	}
	logReplica := make([]int32, 0)
	winNum := len(replicas) / windowSize
	for i := 0; i < winNum; i++ {
		left := i * windowSize
		winMax := utils.Max(replicas[left : left+windowSize]...)
		for j := left; j < left+windowSize; j++ {
			logReplica = append(logReplica, winMax)
		}
	}
	if len(logReplica) == len(replicas) {
		return logReplica, interval
	}
	winMax := utils.Max(replicas[len(replicas)-len(replicas)%windowSize:]...)
	for i := 0; i < len(replicas)%windowSize; i++ {
		logReplica = append(logReplica, winMax)
	}
	return logReplica, interval
}

func infect(replicas []int32, interval time.Duration, windowSize int) ([]int32, time.Duration) {
	if windowSize == 0 {
		return replicas, interval
	}
	max := utils.Max(replicas...)
	logReplica := make([]int32, len(replicas))
	copy(logReplica, replicas)
	firstWindowMax := utils.Max(replicas[:windowSize]...)
	if firstWindowMax == max {
		for i := 0; i < windowSize; i++ {
			logReplica[i] = firstWindowMax
		}
	}
	for i := windowSize; i < len(replicas); {
		if replicas[i] == max {
			for j := i - 1; j >= 0 && i-j <= windowSize; j-- {
				logReplica[j] = max
			}
			for k := 0; k < windowSize; k++ {
				i++
				if i >= len(replicas) {
					break
				}
				logReplica[i] = max
			}
		} else {
			i++
		}
	}
	return logReplica, interval
}
func aggregationBy30(replicas []int32, interval time.Duration) ([]int32, time.Duration) {
	block := 1
	sleepTime := interval
	//find a max replica in an interval of at least 30s
	if interval < 30*time.Second {
		block = int(30 * time.Second / interval)
		sleepTime = 30 * time.Second
	}
	logReplica := make([]int32, 0)
	for i := 0; i < len(replicas); {
		blockEnd := utils.Min(len(replicas), i+block)
		targetReplica := utils.Max(replicas[i:blockEnd]...)
		logReplica = append(logReplica, targetReplica)
		i += block
	}
	return logReplica, sleepTime
}
