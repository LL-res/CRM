package k8s

import (
	"context"
	"fmt"
	"github.com/prometheus/client_golang/api"
	"github.com/prometheus/common/model"
	"log"
	"testing"
	"time"

	promV1 "github.com/prometheus/client_golang/api/prometheus/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestClient_GetReplica(t *testing.T) {
	deploy, err := GlobalClient.ClientSet.AppsV1().Deployments("default").Get(
		context.Background(),
		"my-app-deployment",
		v1.GetOptions{},
	)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(*deploy.Spec.Replicas)
	n, err := GlobalClient.GetReplica("default", autoscalingv2.CrossVersionObjectReference{
		Kind:       "Deployment",
		Name:       "my-app-deployment",
		APIVersion: "apps",
	})
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(n)

}
func TestClient_NewScaleClient(t *testing.T) {
	scale, err := GlobalClient.NewScaleClient()
	if err != nil {
		t.Error(err)
		return
	}
	obj, err := scale.Scales("default").Get(context.Background(), schema.GroupResource{
		Group:    "apps",
		Resource: "Deployment",
	}, "my-app-deployment", v1.GetOptions{})
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(obj.Spec.Replicas)
}
func TestClient_SetReplica(t *testing.T) {
	replica := 10
	start := time.Now()
	err := GlobalClient.SetReplica("default", autoscalingv2.CrossVersionObjectReference{
		Kind:       "Deployment",
		Name:       "diagnosis-system-backend",
		APIVersion: "apps/v1",
	}, int32(replica))
	if err != nil {
		log.Println(err)
		return
	}

	client, err := api.NewClient(api.Config{
		Address: "http://192.168.67.2/prometheus",
	})
	if err != nil {
		log.Fatalf("Error creating client: %v\n", err)
	}

	v1api := promV1.NewAPI(client)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	query := "sum(up{service=\"diagnosis-backend-service\"})"
Loop:
	for {
		result, warnings, err := v1api.Query(ctx, query, time.Now())
		if err != nil {
			log.Fatalf("Error querying Prometheus: %v\n", err)
		}
		if len(warnings) > 0 {
			fmt.Printf("Warnings: %v\n", warnings)
		}
		vectorVal, ok := result.(model.Vector)
		if !ok {
			log.Fatalln("Query result is not a vector")
		}
		for _, sample := range vectorVal {
			if int(sample.Value) == replica {
				break Loop
			}
			//fmt.Printf("Timestamp: %v Value: %v\n", sample.Timestamp, sample.Value)
		}
	}
	end := time.Now()
	fmt.Println(end.Sub(start).Seconds())
}
func TestMain(m *testing.M) {
	err := NewClient()
	if err != nil {
		log.Panic(err)
		return
	}

	m.Run()
}
