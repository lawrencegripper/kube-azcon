package azureProviders

import (
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
)

func StringPointer(i string) *string { return &i }
func Int32Pointer(i int32) *int32    { return &i }
func Int64Pointer(i int64) *int64    { return &i }
func IntPointer(i int) *int          { return &i }

// NewServicePrincipalTokenFromCredentials creates a new ServicePrincipalToken using values of the
// passed credentials map.
func NewServicePrincipalTokenFromCredentials(c ARMConfig, scope string) (*adal.ServicePrincipalToken, error) {
	oauthConfig, err := adal.NewOAuthConfig(azure.PublicCloud.ActiveDirectoryEndpoint, c.TenantID)
	if err != nil {
		panic(err)
	}
	return adal.NewServicePrincipalToken(*oauthConfig, c.ClientID, c.ClientSecret, scope)
}