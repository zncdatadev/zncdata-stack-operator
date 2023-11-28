/*
Copyright 2023 zncdata-labs.

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

// DatabaseConnectionSpec defines the desired state of DatabaseConnection
type DatabaseConnectionSpec struct {
	Provider *ProviderSpec `json:"provider,omitempty"`
}

type ProviderSpec struct {
	// +kubebuilder:validation:Enum=mysql;postgres
	// +kubebulider:default=postgres
	Driver     string          `json:"driver,omitempty"`
	Host       string          `json:"host,omitempty"`
	Port       int             `json:"port,omitempty"`
	SSL        bool            `json:"ssl,omitempty"`
	Credential *CredentialSpec `json:"credential,omitempty"`
}

type CredentialSpec struct {
	ExistSecret string `json:"existSecret,omitempty"`
	Username    string `json:"username,omitempty"`
	Password    string `json:"password,omitempty"`
}

// DatabaseConnectionStatus defines the observed state of DatabaseConnection
type DatabaseConnectionStatus struct {
	// +kubebuilder:validation:Optional
	Conditions []metav1.Condition `json:"condition,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// DatabaseConnection is the Schema for the databaseconnections API
type DatabaseConnection struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DatabaseConnectionSpec   `json:"spec,omitempty"`
	Status DatabaseConnectionStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DatabaseConnectionList contains a list of DatabaseConnection
type DatabaseConnectionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DatabaseConnection `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DatabaseConnection{}, &DatabaseConnectionList{})
}
