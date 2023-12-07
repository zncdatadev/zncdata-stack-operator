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
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DatabaseConnectionSpec defines the desired state of DatabaseConnection
type DatabaseConnectionSpec struct {
	Provider *ProviderSpec `json:"provider,omitempty"`
	Default  bool          `json:"default,omitempty"`
}

type ProviderSpec struct {
	// +kubebuilder:validation:Enum=mysql;postgres
	// +kubebulider:default=postgres
	Driver     string                            `json:"driver,omitempty"`
	Host       string                            `json:"host,omitempty"`
	Port       int                               `json:"port,omitempty"`
	SSL        bool                              `json:"ssl,omitempty"`
	Credential *DatabaseConnectionCredentialSpec `json:"credential,omitempty"`
}

type DatabaseConnectionCredentialSpec struct {
	ExistSecret string `json:"existingSecret,omitempty"`
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

// SetStatusCondition updates the status condition using the provided arguments.
// If the condition already exists, it updates the condition; otherwise, it appends the condition.
// If the condition status has changed, it updates the condition's LastTransitionTime.
func (r *DatabaseConnection) SetStatusCondition(condition metav1.Condition) {
	// if the condition already exists, update it
	existingCondition := apimeta.FindStatusCondition(r.Status.Conditions, condition.Type)
	if existingCondition == nil {
		condition.ObservedGeneration = r.GetGeneration()
		condition.LastTransitionTime = metav1.Now()
		r.Status.Conditions = append(r.Status.Conditions, condition)
	} else if existingCondition.Status != condition.Status || existingCondition.Reason != condition.Reason || existingCondition.Message != condition.Message {
		existingCondition.Status = condition.Status
		existingCondition.Reason = condition.Reason
		existingCondition.Message = condition.Message
		existingCondition.ObservedGeneration = r.GetGeneration()
		existingCondition.LastTransitionTime = metav1.Now()
	}
}

// InitStatusConditions initializes the status conditions to the provided conditions.
func (r *DatabaseConnection) InitStatusConditions() {
	r.Status.Conditions = []metav1.Condition{}
	r.SetStatusCondition(metav1.Condition{
		Type:               ConditionTypeProgressing,
		Status:             metav1.ConditionTrue,
		Reason:             ConditionReasonPreparing,
		Message:            "DatabaseConnection is preparing",
		ObservedGeneration: r.GetGeneration(),
		LastTransitionTime: metav1.Now(),
	})
	r.SetStatusCondition(metav1.Condition{
		Type:               ConditionTypeAvailable,
		Status:             metav1.ConditionFalse,
		Reason:             ConditionReasonPreparing,
		Message:            "DatabaseConnection is preparing",
		ObservedGeneration: r.GetGeneration(),
		LastTransitionTime: metav1.Now(),
	})
}

func (r *DatabaseConnection) IsAvailable() bool {
	if cond := apimeta.FindStatusCondition(r.Status.Conditions, ConditionTypeAvailable); cond != nil && cond.Status == metav1.ConditionTrue && string(cond.Status) == ConditionReasonRunning {
		return true
	}
	return false
}
