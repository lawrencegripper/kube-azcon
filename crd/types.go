package crd

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (v *AzureResource) GetObjectKind() schema.ObjectKind {
	return schema.EmptyObjectKind
}

type AzureResourceList struct {
	v1.TypeMeta `json:",inline"`
	v1.ListMeta `json:"metadata"`
	Items           []AzureResource `json:"items"`
}

type AzureResource struct {
	v1.TypeMeta   `json:",inline"`
	v1.ObjectMeta `json:"metadata"`
	Spec          AzureResourceSpec   `json:"spec"`
	Status        AzureResourceStatus `json:"status,omitempty"`
}

type AzureResourceSpec struct {
	resourceProvider string
	location         string
}

type AzureResourceStatus struct {
	provisioned bool
	errored     bool
	message     string
}
