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

// S3ConnectionSpec defines the desired state of S3Connection
type S3ConnectionSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// +kubebuilder:validation:Required
	S3Credential *S3Credential `json:"credential,omitempty"`
}

type S3Credential struct {
	// +kubebuilder:validation:Optional
	ExistingSecret string `json:"existingSecret,omitempty"`
	AccessKey      string `json:"accessKey,omitempty"`
	SecretKey      string `json:"secretKey,omitempty"`
	Endpoint       string `json:"endpoint,omitempty"`
	Region         string `json:"region,omitempty"`
	SSL            bool   `json:"ssl,omitempty"`
}

// S3ConnectionStatus defines the observed state of S3Connection
type S3ConnectionStatus struct {
	// +kubebuilder:validation:Optional
	Conditions []metav1.Condition `json:"condition,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// S3Connection is the Schema for the s3connections API
type S3Connection struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   S3ConnectionSpec   `json:"spec,omitempty"`
	Status S3ConnectionStatus `json:"status,omitempty"`
}

func (s *S3Connection) SetStatusCondition(condition metav1.Condition) {
	// if the condition already exists, update it
	existingCondition := apimeta.FindStatusCondition(s.Status.Conditions, condition.Type)
	if existingCondition == nil {
		condition.ObservedGeneration = s.GetGeneration()
		condition.LastTransitionTime = metav1.Now()
		s.Status.Conditions = append(s.Status.Conditions, condition)
	} else if existingCondition.Status != condition.Status || existingCondition.Reason != condition.Reason || existingCondition.Message != condition.Message {
		existingCondition.Status = condition.Status
		existingCondition.Reason = condition.Reason
		existingCondition.Message = condition.Message
		existingCondition.ObservedGeneration = s.GetGeneration()
		existingCondition.LastTransitionTime = metav1.Now()
	}
}

func (s *S3Connection) IsAvailable() bool {
	if cond := apimeta.FindStatusCondition(s.Status.Conditions, ConditionTypeAvailable); cond != nil && cond.Status == metav1.ConditionTrue && string(cond.Status) == ConditionReasonRunning {
		return true
	}
	return false
}

func (s *S3Connection) InitStatusConditions() {
	s.Status.Conditions = []metav1.Condition{}
	s.SetStatusCondition(metav1.Condition{
		Type:               ConditionTypeProgressing,
		Status:             metav1.ConditionTrue,
		Reason:             ConditionReasonPreparing,
		Message:            "s3Connection is preparing",
		ObservedGeneration: s.GetGeneration(),
		LastTransitionTime: metav1.Now(),
	})
	s.SetStatusCondition(metav1.Condition{
		Type:               ConditionTypeAvailable,
		Status:             metav1.ConditionFalse,
		Reason:             ConditionReasonPreparing,
		Message:            "s3Connection is preparing",
		ObservedGeneration: s.GetGeneration(),
		LastTransitionTime: metav1.Now(),
	})
}

//+kubebuilder:object:root=true

// S3ConnectionList contains a list of S3Connection
type S3ConnectionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []S3Connection `json:"items"`
}

func init() {
	SchemeBuilder.Register(&S3Connection{}, &S3ConnectionList{})
}
