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

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// Source describes the location of the template source in the same namespace
type Source struct {
	// Name identifies the ConfigMap in the same Namespace as the Template resource that will be used as the source of the template.
	Name string `json:"name"`

	// +kubebuilder:validation:Optional

	// Keys identifies the items in the ConfigMap that will be treated as template sources. If empty all keys will be used.
	Keys []string `json:"keys,omitempty"`
}

// Input describes a source of variables that will be supplied to the template during rendering
type Input struct {
	// Ref identifies the source of the input.
	Name string `json:"name"`

	// +kubebuilder:validation:Enum=secret;configmap
	// +kubebuilder:default=secret

	// Type identifies the type of the input. e.g. configmap, secret
	Type string `json:"type"`

	// +kubebuilder:validation:Optional

	// keys identifies the items in the input that will be treated as input variables. If empty all keys will be used.
	Keys []string `json:"keys,omitempty"`
}

// TemplateSpec defines the desired state of Template
type TemplateSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Source describes where to find the source template. This can be any ConfigMap in the same namespace as this resource.
	Source Source `json:"source"`

	// Inputs describes the set of secrets and configmaps to add to the template variables.
	Inputs []Input `json:"inputs"`

	// Output is the name of the secret to store the rendered template(s) output in.
	Output string `json:"output"`
}

// TemplateStatus defines the observed state of Template
type TemplateStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Template is the Schema for the templates API
type Template struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TemplateSpec   `json:"spec,omitempty"`
	Status TemplateStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TemplateList contains a list of Template
type TemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Template `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Template{}, &TemplateList{})
}
