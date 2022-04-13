/*
Copyright 2022.

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ProjectSpec defines the desired state of Project
type ProjectSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// ResourceQuota specification
	ResourceQuota *ResourceQuota `json:"resourceQuota"`
	// ResourceQuota specification
	LimitRange *LimitRange `json:"limitRange"`
}
type ResourceQuota struct {

	// Requests for CPU
	// +kubebuilder:default:="2.5"
	RequestsCpu string `json:"requests.cpu"`

	// Requests for Memory
	// +kubebuilder:default:="1G"
	RequestsMemory string `json:"requests.memory"`

	// Limits for CPU
	// +kubebuilder:default:="2"
	LimitsCpu string `json:"limits.cpu"`

	// Limits for Memory
	// +kubebuilder:default:="4G"
	LimitsMemory string `json:"limits.memory"`
}

type LimitRange struct {

	// Maksimum for CPU
	// +kubebuilder:default:="30G"
	MaxCpu string `json:"maxCpu"`

	// Maksimum for Memory
	// +kubebuilder:default:="20Gi"
	MaxMemory string `json:"maxMemory"`

	// Minimum for CPU
	// +kubebuilder:default:="50m"
	MinCpu string `json:"minCpu"`

	// Minimum for Memory
	// +kubebuilder:default:="50Mi"
	MinMemory string `json:"minMemory"`

	// Default limit for Cpu
	// +kubebuilder:default:="1000m"
	DefaultCpu string `json:"defaultCpu"`

	// Default limit for Memory
	// +kubebuilder:default:="1000Mi"
	DefaultMemory string `json:"defaultMemory"`

	// Default request for Cpu
	// +kubebuilder:default:="100m"
	DefaultRequestCpu string `json:"defaultRequestCpu"`

	// Default request for Memory
	// +kubebuilder:default:="100m"
	DefaultRequestmemory string `json:"defaultRequestMemory"`
}

// ProjectStatus defines the observed state of Project
type ProjectStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Project is the Schema for the projects API
type Project struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProjectSpec   `json:"spec,omitempty"`
	Status ProjectStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ProjectList contains a list of Project
type ProjectList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Project `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Project{}, &ProjectList{})
}
