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

// ProjectRoleSpec defines the desired state of ProjectRole
type ProjectRoleSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// Project name
	ProjectName string `json:"projectName,omitempty"`
	// Role names array
	Roles []string `json:"roles,omitempty"`
}

// ProjectRoleStatus defines the observed state of ProjectRole
type ProjectRoleStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName=projrole;prrole

// ProjectRole is the Schema for the projectroles API
type ProjectRole struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProjectRoleSpec   `json:"spec,omitempty"`
	Status ProjectRoleStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ProjectRoleList contains a list of ProjectRole
type ProjectRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ProjectRole `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ProjectRole{}, &ProjectRoleList{})
}
