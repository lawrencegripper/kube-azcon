package sql

import (
	// "net/http"

	"github.com/Azure/azure-sdk-for-go/arm/sql"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/golang/glog"
	"github.com/lawrencegripper/kube-azureresources/azureProviders"
	"fmt"
)

// NewServicePrincipalTokenFromCredentials creates a new ServicePrincipalToken using values of the
// passed credentials map.
func newServicePrincipalTokenFromCredentials(c azureProviders.ARMConfig, scope string) (*adal.ServicePrincipalToken, error) {
	oauthConfig, err := adal.NewOAuthConfig(azure.PublicCloud.ActiveDirectoryEndpoint, c.TenantID)
	if err != nil {
		panic(err)
	}
	return adal.NewServicePrincipalToken(*oauthConfig, c.ClientID, c.ClientSecret, scope)
}

func Deploy(serverName, databaseName, location string) (string, error) {
	c, err := azureProviders.GetConfigFromEnv()

	if err != nil {
		glog.Fatalf("Failed to get azure configuration from environment: %v", err)
	}

	spt, err := newServicePrincipalTokenFromCredentials(c, azure.PublicCloud.ResourceManagerEndpoint)

	if err != nil {
		glog.Fatalf("Failed creating service principal: %v", err)
	}

	auth := autorest.NewBearerAuthorizer(spt)

	sqlServerClient := sql.NewServersClient(c.SubscriptionID)
	sqlServerClient.Authorizer = auth

	server, err := sqlServerClient.Get(c.ResourceGroup, serverName)
	// if err != nil && err.(*autorest.DetailedError).StatusCode == http.StatusNotFound {
	// 	sqlServerClient.CreateOrUpdate(
	// 		c.ResourceGroup,
	// 		serverName,
	// 		sql.Server{
	// 			ServerProperties: sql.ServerProperties{
	// 				AdministratorLogin:         "lawrence",
	// 				AdministratorLoginPassword: "140912384alsdkfaosidurer341",
	// 				Kind: &"Microsoft.DBforPostgreSQL/servers",
	// 			},
	// 		})
	// }

	fmt.Print(server.Name)

	// sqlDbClient := sql.NewDatabasesClient(c.SubscriptionID)
	// sqlDbClient.Authorizer = auth

	// sqlClient.CreateOrUpdate(c.ResourceGroup, serverName, databaseName)

	// sqlClient.Get(os.Getenv("AZURE_RESOURCE_GROUP"), )

	return "hello", nil
}
