package scaler

import (
	"fmt"
	"testing"
	"time"
)

// 集群不开启时也可以进行测试

func TestSlopeWindow_Infect(t *testing.T) {
	infector := SlopeWindow{}
	replicas1 := make([]int32, 0)
	for i := 0; i < 5; i++ {
		replicas1 = append(replicas1, int32(i))
	}
	for i := 5; i > 0; i-- {
		replicas1 = append(replicas1, int32(i))
	}
	fmt.Println(infector.Aggregate(replicas1, time.Second, 2))
}
