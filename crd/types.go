package crd

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"encoding/json"
	"github.com/lawrencegripper/kube-azureresources/azureProviders"
	"time"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	AzureResourceKind = "AzureResource"
	AzureResourceType = "azureresources"
)

func (v *AzureResource) GetObjectKind() schema.ObjectKind {
	return schema.EmptyObjectKind
}

type AzureResourceList struct {
	v1.TypeMeta `json:",inline"`
	v1.ListMeta `json:"metadata"`
	Items       []AzureResource `json:"items"`
}

type AzureResource struct {
	v1.TypeMeta   `json:",inline"`
	v1.ObjectMeta `json:"metadata"`
	Spec          AzureResourceSpec   `json:"spec"`
	Status        AzureResourceStatus `json:"status,omitempty"`
}

type AzureResourceSpec struct {
	ResourceProvider string `json:"resourceProvider"`
	Location         string `json:"location"`
}

type AzureResourceStatus struct {
	ProvisioningStatus string `json:"provisioningStatus"`
	LastChecked        time.Time `json:"lastChecked"`
	Output			   azureProviders.Output `json:"ouput"`
}


func (a *AzureResource) AsUnstructured() (*unstructured.Unstructured, error) {
	a.TypeMeta.Kind = AzureResourceKind
	a.TypeMeta.APIVersion = versionedGroupName.Group + "/" + versionedGroupName.Version
	b, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}
	var r unstructured.Unstructured
	if err := json.Unmarshal(b, &r.Object); err != nil {
		return nil, err
	}
	return &r, nil
}