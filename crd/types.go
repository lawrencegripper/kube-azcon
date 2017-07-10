package crd

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/api/unversioned"
)

func (v *AzureResource) GetObjectKind() unversioned.ObjectKind {
	return unversioned.EmptyObjectKind
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
