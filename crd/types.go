package crd

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (v *AzureResource) GetObjectKind() schema.ObjectKind {
	return schema.EmptyObjectKind
}

type AzureResource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              AzureResourceSpec   `json:"spec"`
	Status            AzureResourceStatus `json:"status,omitempty"`
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
