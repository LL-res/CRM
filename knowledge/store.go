package knowledge

import (
	"k8s.io/apimachinery/pkg/types"
)

var clusterKnowledge ClusterKnowledge

func GetLocalKnowledge(name types.NamespacedName) *LocalKnowledge {
	if nil == clusterKnowledge {
		clusterKnowledge = make(map[types.NamespacedName]*LocalKnowledge)
	}
	if nil == clusterKnowledge[name] {
		h := new(LocalKnowledge)
		h.Init()
		clusterKnowledge[name] = h
	}
	return clusterKnowledge[name]
}
