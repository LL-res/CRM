package k8s

import (
	"context"
	"log"
	"sync"

	autoscalingv2 "k8s.io/api/autoscaling/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	cached "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/scale"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	GlobalClient *Client
	once         sync.Once
)

type Client struct {
	Config      *rest.Config
	ClientSet   *kubernetes.Clientset
	ScaleGetter scale.ScalesGetter
}

func NewClient() error {
	var err error
	once.Do(func() {
		err = newClient()
	})
	return err
}
func newClient() error {
	conf, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		log.Println(err)
		return err
	}
	clientSet, err := kubernetes.NewForConfig(conf)
	if err != nil {
		log.Println(err)
		return err
	}
	apiGroupResources, err := restmapper.GetAPIGroupResources(clientSet)
	if err != nil {
		log.Println(err)
		return err
	}
	mapper := restmapper.NewDiscoveryRESTMapper(apiGroupResources)
	resolver := scale.NewDiscoveryScaleKindResolver(clientSet)

	scalesGetter, err := scale.NewForConfig(conf, mapper, dynamic.LegacyAPIPathResolverFunc, resolver)
	if err != nil {
		log.Println(err)
		return err
	}
	GlobalClient = &Client{
		Config:      conf,
		ClientSet:   clientSet,
		ScaleGetter: scalesGetter,
	}
	return nil
}
func (c *Client) NewScaleClient() (scale.ScalesGetter, error) {
	scaleClient, err := scale.NewForConfig(c.Config,
		restmapper.NewDeferredDiscoveryRESTMapper(cached.NewMemCacheClient(c.ClientSet.Discovery())),
		dynamic.LegacyAPIPathResolverFunc,
		scale.NewDiscoveryScaleKindResolver(c.ClientSet.Discovery()))
	if err != nil {
		return nil, err
	}
	return scaleClient, nil

}
func (c *Client) GetReplica(namespace string, scaleTargetRef autoscalingv2.CrossVersionObjectReference) (int32, error) {
	gvk := schema.FromAPIVersionAndKind(scaleTargetRef.APIVersion, "")
	scaleObj, err := c.ScaleGetter.Scales(namespace).Get(context.TODO(), schema.GroupResource{
		Group:    gvk.Group,
		Resource: scaleTargetRef.Kind,
	}, scaleTargetRef.Name, metav1.GetOptions{})
	if err != nil {
		return 0, err
	}
	return scaleObj.Spec.Replicas, nil
}
func (c *Client) SetReplica(namespace string, scaleTargetRef autoscalingv2.CrossVersionObjectReference, replica int32) error {
	gvk := schema.FromAPIVersionAndKind(scaleTargetRef.APIVersion, "")
	scaleObj, err := GlobalClient.ScaleGetter.Scales(namespace).Get(context.TODO(), schema.GroupResource{
		Group:    gvk.Group,
		Resource: scaleTargetRef.Kind,
	}, scaleTargetRef.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	scaleObj.Spec.Replicas = replica
	_, err = GlobalClient.ScaleGetter.Scales(namespace).Update(context.TODO(), schema.GroupResource{
		Group:    gvk.Group,
		Resource: scaleTargetRef.Kind,
	}, scaleObj, metav1.UpdateOptions{})

	return err
}
