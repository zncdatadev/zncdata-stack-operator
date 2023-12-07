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

const DatabaseFinalizer = "database.finalizers.stack.zncdata.net"

// DatabaseSpec defines the desired state of Database
type DatabaseSpec struct {
	Name       string                  `json:"name,omitempty"`
	Reference  string                  `json:"reference,omitempty"`
	Credential *DatabaseCredentialSpec `json:"credential,omitempty"`
}

type DatabaseCredentialSpec struct {
	ExistSecret string `json:"existingSecret,omitempty"`
	Username    string `json:"username,omitempty"`
	Password    string `json:"password,omitempty"`
}

// DatabaseStatus defines the observed state of Database
type DatabaseStatus struct {
	// +kubebuilder:validation:Optional
	Conditions []metav1.Condition `json:"condition,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Database is the Schema for the databases API
type Database struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DatabaseSpec   `json:"spec,omitempty"`
	Status DatabaseStatus `json:"status,omitempty"`
}

func (in *Database) GetNameWithSuffix(s string) string {
	return in.Name + "-" + s
}

func (in *Database) SetStatusCondition(condition metav1.Condition) {
	// if the condition already exists, update it
	existingCondition := apimeta.FindStatusCondition(in.Status.Conditions, condition.Type)
	if existingCondition == nil {
		condition.ObservedGeneration = in.GetGeneration()
		condition.LastTransitionTime = metav1.Now()
		in.Status.Conditions = append(in.Status.Conditions, condition)
	} else if existingCondition.Status != condition.Status || existingCondition.Reason != condition.Reason || existingCondition.Message != condition.Message {
		existingCondition.Status = condition.Status
		existingCondition.Reason = condition.Reason
		existingCondition.Message = condition.Message
		existingCondition.ObservedGeneration = in.GetGeneration()
		existingCondition.LastTransitionTime = metav1.Now()
	}
}

func (in *Database) IsAvailable() bool {
	if cond := apimeta.FindStatusCondition(in.Status.Conditions, ConditionTypeAvailable); cond != nil && cond.Status == metav1.ConditionTrue && string(cond.Status) == ConditionReasonRunning {
		return true
	}
	return false
}

//+kubebuilder:object:root=true

// DatabaseList contains a list of Database
type DatabaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Database `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Database{}, &DatabaseList{})
}
