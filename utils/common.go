package utils

import (
	"context"
	"fmt"
	"github.com/LL-res/CRM/common/consts"
	"k8s.io/apimachinery/pkg/types"
	"strings"
	"time"

	"golang.org/x/exp/constraints"
)

func Max[T constraints.Ordered](x ...T) T {
	if len(x) == 1 {
		return x[0]
	}
	max := x[0]
	for i := 1; i < len(x); i++ {
		if x[i] > max {
			max = x[i]
		}
	}
	return max
}

func Min[T constraints.Ordered](x ...T) T {
	if len(x) == 1 {
		return x[0]
	}
	min := x[0]
	for i := 1; i < len(x); i++ {
		if x[i] < min {
			min = x[i]
		}
	}
	return min
}
func GetWithModelKey(NoModelKey, model string) string {
	return fmt.Sprintf("%s$%s", NoModelKey, model)
}
func GetNoModelKey(withModelKey string) string {
	strs := strings.Split(withModelKey, "$")
	return strings.Join(strs[:len(strs)-1], "$")
}
func GetModelType(withModelType string) string {
	strs := strings.Split(withModelType, "$")
	return strs[len(strs)-1]
}
func MulSlice[T constraints.Float | constraints.Integer](k T, nums []T) {
	for i := range nums {
		nums[i] *= k
	}
}
func AddSlice[T constraints.Float | constraints.Integer](nums ...[]T) []T {
	res := make([]T, len(nums[0]))
	for _, num := range nums {
		for j, v := range num {
			res[j] += v
		}
	}
	return res
}
func GetInstanceName(ctx context.Context) types.NamespacedName {
	return types.NamespacedName{
		Namespace: ctx.Value(consts.NAMESPACE).(string),
		Name:      ctx.Value(consts.NAME).(string),
	}
}
func GetMSELoss(x, y []float64) float64 {
	totalLoss := 0.0
	for i := range x {
		totalLoss += (x[i] - y[i]) * (x[i] - y[i])
	}
	return totalLoss / float64(len(x))
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
func Caliberate(interval int) *time.Ticker {
	msCalibrator := time.NewTicker(time.Millisecond)
	for tt := range msCalibrator.C {
		if tt.UnixMilli()%1000 == 0 {
			break
		}
	}
	msCalibrator.Stop()
	SecCalibrator := time.NewTicker(time.Second)
	for tt := range SecCalibrator.C {
		if tt.Second()%interval == 0 {
			break
		}
	}
	SecCalibrator.Stop()
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	return ticker
}
func CaliberateTime(x time.Time, interval time.Duration) time.Time {
	overSec := x.Second() % int(interval/time.Second)
	overMilli := x.UnixMilli() % 1000
	y := x.Add(-time.Duration(overSec) * time.Second)
	z := y.Add(-time.Duration(overMilli) * time.Millisecond)
	return z
}
