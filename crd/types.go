package crd

import (
	"github.com/lawrencegripper/kube-azcon/models"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"encoding/json"
	"time"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	AzureResourceKind = "AzureResource"
	AzureResourceType = "azureresources"
	TagKubernetesResourceLink = "KubernetesLink"
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
	Output			   models.Output `json:"ouput"`
}
func stringPointer(i string) *string { return &i }

func (a *AzureResource) GenerateAzureTags() map[string]*string {
	return map[string]*string{
		"CreatedBy": stringPointer("kube-azcon"),
		"KubernetesResourceName": &a.Name,
		TagKubernetesResourceLink: &a.SelfLink,
		"KubernetesResourceVersion": &a.ResourceVersion,
	}
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

func AzureResourceFromUnstructured(r *unstructured.Unstructured) (*AzureResource, error) {
	b, err := json.Marshal(r.Object)
	if err != nil {
		return nil, err
	}
	var a AzureResource
	if err := json.Unmarshal(b, &a); err != nil {
		return nil, err
	}
	a.TypeMeta.Kind = AzureResourceKind
	a.TypeMeta.APIVersion = versionedGroupName.Group + "/" + versionedGroupName.Version
	return &a, nil
}