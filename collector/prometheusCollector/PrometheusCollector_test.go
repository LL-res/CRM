package prometheusCollector

import (
	"fmt"
	"github.com/LL-res/CRM/common/key"
	"github.com/LL-res/CRM/common/log"
	"testing"
	"time"
)

func TestGetOrNewFacade(t *testing.T) {
	newFacade, err := GetOrNewFacade(3*time.Second, 2000, "http://192.168.67.2/prometheus", 4000)
	if err != nil {
		log.Logger.Error(err, "")
		return
	}
	nmk := key.NoModelKey{
		MetricName:  "http_request",
		MetricUnit:  "perPod",
		MetricQuery: "sum(http_requests_total - http_requests_total offset 3s)",
	}
	newFacade.CreateCollector(nmk)
	time.Sleep(500 * time.Millisecond)

	fmt.Println(newFacade.GetMetricFromCollector(nmk, 100))
}
