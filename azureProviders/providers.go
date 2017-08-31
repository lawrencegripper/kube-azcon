package azureProviders

import (
	"github.com/lawrencegripper/kube-azcon/models"
	"github.com/lawrencegripper/kube-azcon/crd"
)

type Provider interface {
	CreateOrUpdate(azConfig ARMConfig, azRes crd.AzureResource) (models.Output, error)
}