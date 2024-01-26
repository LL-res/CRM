package BO

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +kubebuilder:object:root=false
type Metric struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Target            string `json:"target"`
	//值范围为 [0,100]
	Weight int32  `json:"weight"`
	Name   string `json:"name"`
	Unit   string `json:"unit"`
	Query  string `json:"query"`
}

// +kubebuilder:object:root=false
type Model struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	//模型的类型，如GRU,Holt-Winter等
	Type      string `json:"type,omitempty"`
	NeedTrain bool   `json:"needTrain,omitempty"`
	//if NeedTrain is true then PreTrained show if the .pt file is first provided
	PreTrained bool `json:"preTrained,omitempty"`
	TrainSize  int  `json:"trainSize,omitempty"`
	// if NeedTrain is true then UpdateInterval show when to update the model
	UpdateInterval int    `json:"updateInterval,omitempty"`
	LookBackward   int    `json:"lookBackward"`
	Debug          bool   `json:"debug,omitempty"`
	SourceImplURL  string `json:"sourceImplURL,omitempty"`
	Command        string `json:"command,omitempty"`
	//模型中的超参数，或独有配置
	Attr map[string]string `json:"attr"`
}
