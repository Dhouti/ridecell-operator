/*
Copyright 2019-2020 Ridecell, Inc.

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

// RabbitmqPermission defines a single user permissions entry.
type RabbitmqPermission struct {
	// Vhost this applies to.
	Vhost string `json:"vhost"`
	// Configuration permissions.
	Configure string `json:"configure"`
	// Write permissions.
	Write string `json:"write"`
	// Read permissions.
	Read string `json:"read"`
}

// RabbitmqUserSpec defines the desired state of RabbitmqUser
type RabbitmqUserSpec struct {
	Username    string               `json:"username"`
	Tags        string               `json:"tags,omitempty"`
	Permissions []RabbitmqPermission `json:"permissions,omitempty"`
	// TODO TopicPermissions
	Connection RabbitmqConnection `json:"connection,omitempty"`
}

// RabbitmqUserStatus defines the observed state of RabbitmqUser
type RabbitmqUserStatus struct {
	Status     string                   `json:"status"`
	Message    string                   `json:"message"`
	Connection RabbitmqStatusConnection `json:"connection,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RabbitmqUser is the Schema for the rabbitmqusers API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type RabbitmqUser struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RabbitmqUserSpec   `json:"spec,omitempty"`
	Status RabbitmqUserStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RabbitmqUserList contains a list of RabbitmqUser
type RabbitmqUserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RabbitmqUser `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RabbitmqUser{}, &RabbitmqUserList{})
}
