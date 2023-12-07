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

// RedisConnectionSpec defines the desired state of RedisConnection
type RedisConnectionSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Host     string `json:"host,omitempty"`
	Port     string `json:"port,omitempty"`
	Password string `json:"password,omitempty"`
}

// RedisConnectionStatus defines the observed state of RedisConnection
type RedisConnectionStatus struct {
	// +kubebuilder:validation:Optional
	Conditions []metav1.Condition `json:"condition,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// RedisConnection is the Schema for the redisconnections API
type RedisConnection struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RedisConnectionSpec   `json:"spec,omitempty"`
	Status RedisConnectionStatus `json:"status,omitempty"`
}

func (r *RedisConnection) SetStatusCondition(condition metav1.Condition) {
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

func (r *RedisConnection) IsAvailable() bool {
	if cond := apimeta.FindStatusCondition(r.Status.Conditions, ConditionTypeAvailable); cond != nil && cond.Status == metav1.ConditionTrue && string(cond.Status) == ConditionReasonRunning {
		return true
	}
	return false
}

func (r *RedisConnection) InitStatusConditions() {
	r.Status.Conditions = []metav1.Condition{}
	r.SetStatusCondition(metav1.Condition{
		Type:               ConditionTypeProgressing,
		Status:             metav1.ConditionTrue,
		Reason:             ConditionReasonPreparing,
		Message:            "redisConnection is preparing",
		ObservedGeneration: r.GetGeneration(),
		LastTransitionTime: metav1.Now(),
	})
	r.SetStatusCondition(metav1.Condition{
		Type:               ConditionTypeAvailable,
		Status:             metav1.ConditionFalse,
		Reason:             ConditionReasonPreparing,
		Message:            "redisConnection is preparing",
		ObservedGeneration: r.GetGeneration(),
		LastTransitionTime: metav1.Now(),
	})
}

//+kubebuilder:object:root=true

// RedisConnectionList contains a list of RedisConnection
type RedisConnectionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RedisConnection `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RedisConnection{}, &RedisConnectionList{})
}
