package azureProviders

import (
	"github.com/lawrencegripper/kube-azureresources/models"
	"github.com/lawrencegripper/kube-azureresources/crd"
)

type Provider interface {
	CreateOrUpdate(azConfig ARMConfig, azRes crd.AzureResource) (models.Output, error)
}