package scaler

import (
	"github.com/LL-res/AOM/utils"
	"log"
	"time"
)

type Aggregator interface {
	Aggregate(replicas []int32, interval time.Duration, windowSize int) ([]int32, time.Duration)
}

type MaxWindow struct{}

type SlopeWindow struct{}

func NewInfector(name string) Aggregator {
	switch name {
	case "MaxWindow":
		return MaxWindow{}
	case "SlopeWindow":
		return SlopeWindow{}
	default:
		log.Println("Unknown infector")
		return SlopeWindow{}
	}
}

func (s SlopeWindow) Aggregate(replicas []int32, interval time.Duration, windowSize int) ([]int32, time.Duration) {
	if windowSize == 0 {
		return replicas, interval
	}
	logReplica := make([]int32, len(replicas))
	copy(logReplica, replicas)
	for i := 1; i < len(replicas); i++ {
		if logReplica[i] > logReplica[i-1] {
			infectTimes := windowSize
			for j := i - 1; j >= 0 && infectTimes > 0; j-- {
				if logReplica[j] < logReplica[i] {
					logReplica[j] = logReplica[i]
				} else {
					break
				}
				infectTimes--
			}
		}
	}
	return logReplica, interval
}

func (w MaxWindow) Aggregate(replicas []int32, interval time.Duration, windowSize int) ([]int32, time.Duration) {
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
