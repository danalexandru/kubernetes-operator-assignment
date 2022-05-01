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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CustomDeploymentSpec defines the desired state of CustomDeployment
type CustomDeploymentSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of CustomDeployment. Edit customdeployment_types.go to remove/update
	// Foo string `json:"foo,omitempty"`

	// Host is where the application is accessible
	Host string `json:"host"`

	// Port is the port where the application is exposed (Default: 8080)
	Port *uint16 `json:"port,omitempty"`

	// Replicas is the number of CustomDeployment replicas (Default: 1)
	Replicas *uint16 `json:"replicas,omitempty"`

	// Image is the container image & tag
	Image Image `json:"image"`
}

// CustomDeploymentStatus defines the observed state of CustomDeployment
type CustomDeploymentStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Deployed bool `json:"deployed"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CustomDeployment is the Schema for the customdeployments API
type CustomDeployment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CustomDeploymentSpec   `json:"spec,omitempty"`
	Status CustomDeploymentStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CustomDeploymentList contains a list of CustomDeployment
type CustomDeploymentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CustomDeployment `json:"items"`
}

// Image defines the Repository, Tag and ImagePullPolicy of the Ingress Controller Image.
type Image struct {
	// The repository of the image.
	Repository string `json:"repository"`
	// The tag (version) of the image.
	Tag string `json:"tag"`
	// The ImagePullPolicy of the image.
	// +kubebuilder:validation:Enum=Never;Always;IfNotPresent
	PullPolicy string `json:"pullPolicy"`
}

func init() {
	SchemeBuilder.Register(&CustomDeployment{}, &CustomDeploymentList{})
}
