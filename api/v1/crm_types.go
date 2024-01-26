/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	"github.com/LL-res/CRM/domain/BO"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CRMSpec defines the desired state of CRM
type CRMSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	ScaleTargetRef autoscalingv2.CrossVersionObjectReference `json:"scaleTargetRef"`
	Collector      Collector                                 `json:"collector"`
	Metrics        map[string]BO.Metric                      `json:"metrics"`
	Models         Models                                    `json:"models"`
	// +kubebuilder:validation:Minimum=0
	// +optional
	MinReplicas int32 `json:"minReplicas"`
	// +kubebuilder:validation:Minimum=1
	// +optional
	MaxReplicas int32 `json:"maxReplicas"`
	// the interval between two time points,unit second
	IntervalDuration int `json:"intervalDuration"`
	// all  models use this as lookFroward

}

// CRMStatus defines the observed state of CRM
type CRMStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Generation int64 `json:"generation"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CRM is the Schema for the crms API
type CRM struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CRMSpec   `json:"spec,omitempty"`
	Status CRMStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CRMList contains a list of CRM
type CRMList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CRM `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CRM{}, &CRMList{})
}

type Collector struct {
	Address        string `json:"address"`
	ScrapeInterval int    `json:"scrapeInterval"`
	BaseOnHistory  int    `json:"baseOnHistory,omitempty"`
	MaxCap         int    `json:"maxCap"`
}

// +kubebuilder:object:root=false
type Models struct {
	LookForward int                   `json:"lookForward"`
	Attr        map[string][]BO.Model `json:"attr"`
}
