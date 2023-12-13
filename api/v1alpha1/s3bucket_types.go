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

const S3BucketFinalizer = "s3bucket.finalizers.stack.zncdata.net"

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// S3BucketSpec defines the desired state of S3Bucket
type S3BucketSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// +kubebuilder:validation:Required
	Reference string `json:"reference,omitempty"`
	// +kubebuilder:validation:Optional
	Credential *S3BucketCredential `json:"credential,omitempty"`
	Name       string              `json:"name,omitempty"`
}

type S3BucketCredential struct {
	// +kubebuilder:validation:Optional
	ExistingSecret string `json:"existingSecret,omitempty"`
}

// S3BucketStatus defines the observed state of S3Bucket
type S3BucketStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Conditions []metav1.Condition `json:"condition,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// S3Bucket is the Schema for the s3buckets API
type S3Bucket struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   S3BucketSpec   `json:"spec,omitempty"`
	Status S3BucketStatus `json:"status,omitempty"`
}

func (s *S3Bucket) InitStatusConditions() {
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

func (s *S3Bucket) SetStatusCondition(condition metav1.Condition) {
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

func (s *S3Bucket) IsAvailable() bool {
	if cond := apimeta.FindStatusCondition(s.Status.Conditions, ConditionTypeAvailable); cond != nil && cond.Status == metav1.ConditionTrue && string(cond.Status) == ConditionReasonRunning {
		return true
	}
	return false
}

//+kubebuilder:object:root=true

// S3BucketList contains a list of S3Bucket
type S3BucketList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []S3Bucket `json:"items"`
}

func init() {
	SchemeBuilder.Register(&S3Bucket{}, &S3BucketList{})
}
