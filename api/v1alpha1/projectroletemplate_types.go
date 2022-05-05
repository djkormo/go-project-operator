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
	v1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ProjectRoleTemplateSpec defines the desired state of ProjectRoleTemplate
type ProjectRoleTemplateSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// Exclude namespaces
	ExcludeNamespaces []string `json:"excludeNamespaces,omitempty"`
	// RBAC Role Rules
	RoleRules []v1.PolicyRule `json:"roleRules,omitempty"`
}

// ProjectRoleTemplateStatus defines the observed state of ProjectRoleTemplate
type ProjectRoleTemplateStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:singular=projectroletemplate
//+kubebuilder:resource:shortName=projroletemplate;projroletemp;prroletemp

// ProjectRoleTemplate is the Schema for the projectroletemplates API
type ProjectRoleTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProjectRoleTemplateSpec   `json:"spec,omitempty"`
	Status ProjectRoleTemplateStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ProjectRoleTemplateList contains a list of ProjectRoleTemplate
type ProjectRoleTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ProjectRoleTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ProjectRoleTemplate{}, &ProjectRoleTemplateList{})
}
