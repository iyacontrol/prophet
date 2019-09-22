/*
Copyright 2017 The Kubernetes Authors.
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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CronHpa is a specification for a Foo resource
type CronHpa struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec CronHpaSpec `json:"spec"`
}

// Cycle is the struct of hour and capacity
type Cycle struct {
	Hour     int32 `json:"hour"`
	Capacity int32 `json:"capacity"`
}

// CronHpaSpec is the spec for a CronHpa resource
type CronHpaSpec struct {
	HpaName string  `json:"hpa_name"`
	Cycles  []Cycle `json:"cycles"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CronHpaList is a list of CronHpa resources
type CronHpaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []CronHpa `json:"items"`
}
