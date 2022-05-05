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
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ProjectNetworkPolicyTemplateSpec defines the desired state of ProjectNetworkPolicyTemplate
type ProjectNetworkPolicyTemplateSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Exclude namespaces
	ExcludeNamespaces []string `json:"excludeNamespaces,omitempty"`
	// Network Policy specitifation
	PolicySpec networkingv1.NetworkPolicySpec `json:"policySpec"`
}

// ProjectNetworkPolicyTemplateStatus defines the observed state of ProjectNetworkPolicyTemplate
type ProjectNetworkPolicyTemplateStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:singular=projectnetworkpolicytemplate
//+kubebuilder:resource:scope=Cluster
//+kubebuilder:resource:shortName=projectnetpoltemplate;projectnetpoltemp;projnetpoltemp

// ProjectNetworkPolicyTemplate is the Schema for the projectnetworkpolicytemplates API
type ProjectNetworkPolicyTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProjectNetworkPolicyTemplateSpec   `json:"spec,omitempty"`
	Status ProjectNetworkPolicyTemplateStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ProjectNetworkPolicyTemplateList contains a list of ProjectNetworkPolicyTemplate
type ProjectNetworkPolicyTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ProjectNetworkPolicyTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ProjectNetworkPolicyTemplate{}, &ProjectNetworkPolicyTemplateList{})
}
