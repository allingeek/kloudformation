/*
Copyright 2018 The KloudFormation authors.

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

// InternetGatewayAttachmentSpec defines the desired state of InternetGatewayAttachment
type InternetGatewayAttachmentSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	InternetGatewayName string `json:"internetGatewayName"`
	VPCName             string `json:"vpcName"`
}

// InternetGatewayAttachmentStatus defines the observed state of InternetGatewayAttachment
type InternetGatewayAttachmentStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// InternetGatewayAttachment is the Schema for the internetgatewayattachments API
// +k8s:openapi-gen=true
type InternetGatewayAttachment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InternetGatewayAttachmentSpec   `json:"spec,omitempty"`
	Status InternetGatewayAttachmentStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// InternetGatewayAttachmentList contains a list of InternetGatewayAttachment
type InternetGatewayAttachmentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []InternetGatewayAttachment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&InternetGatewayAttachment{}, &InternetGatewayAttachmentList{})
}
