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

// IP validation based on https://github.com/metal3-io/ip-address-manager/blob/main/api/v1alpha1/common_types.go

// EndPoint is a specification for a resource ProjectAccessSpec
type EndPoint struct {
	// Endpoint name
	Name string `json:"name"`
	// IP is used for validation of an IP address
	// +kubebuilder:validation:Pattern="((^((([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5]))$)|(^(([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:))$))"
	IP string `json:"ip"`
	// Protocol of endpoint
	Protocol string `json:"protocol"`
	// Port of endpoint
	Port int `json:"port"`
	// Description of endpoint
	Description string `json:"description,omitempty"`
}

// ProjectAccessSpec defines the desired state of ProjectAccess
type ProjectAccessSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	ProjectName string     `json:"projectName,omitempty"`
	Endpoints   []EndPoint `json:"endpoints"`
}

// ProjectAccessStatus defines the observed state of ProjectAccess
type ProjectAccessStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName=projaccess;praccess
//+kubebuilder:printcolumn:name="ProjectName",type=string,JSONPath=`.spec.projectName`

// ProjectAccess is the Schema for the projectaccesses API
type ProjectAccess struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProjectAccessSpec   `json:"spec,omitempty"`
	Status ProjectAccessStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ProjectAccessList contains a list of ProjectAccess
type ProjectAccessList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ProjectAccess `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ProjectAccess{}, &ProjectAccessList{})
}
