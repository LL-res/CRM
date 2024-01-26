package utils

import (
	"fmt"
	"testing"
	"time"
)

func TestAddSlice(t *testing.T) {
	base := [][]int32{{2, 2, 2, 2, 2, 2, 4, 4}}
	res := AddSlice(base...)
	fmt.Println(res)
}
func TestCaliberateTime(t *testing.T) {
	now := time.Now()
	fmt.Println(CaliberateTime(now, 3))
}
