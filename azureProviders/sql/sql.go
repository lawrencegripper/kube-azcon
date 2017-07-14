package sql

import (
	"net/http"

	"fmt"
	"math/rand"
	"time"

	"github.com/Azure/azure-sdk-for-go/arm/sql"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/golang/glog"
	"github.com/lawrencegripper/kube-azureresources/azureProviders"
)

type Config struct {
	ServerName                 string
	DatabaseName               string
	Location                   string
	AdministratorLogin         string
	AdministratorLoginPassword string
	Kind                       string
}

func NewPostgresConfig(serverName, databaseName, location string) Config {
	config := Config{ServerName: serverName, DatabaseName: databaseName, Location: location}
	config.AdministratorLogin = "azurePostgres"
	config.AdministratorLoginPassword = randAlphaNumericSeq(18)
	return config
}

func Deploy(deployConfig Config, azConfig azureProviders.ARMConfig) (string, error) {
	azConfig, err := azureProviders.GetConfigFromEnv()

	if err != nil {
		glog.Fatalf("Failed to get azure configuration from environment: %v", err)
	}

	spt, err := newServicePrincipalTokenFromCredentials(azConfig, azure.PublicCloud.ResourceManagerEndpoint)

	if err != nil {
		glog.Fatalf("Failed creating service principal: %v", err)
	}

	auth := autorest.NewBearerAuthorizer(spt)

	sqlServerClient := sql.NewServersClient(azConfig.SubscriptionID)
	sqlServerClient.Authorizer = auth

	server, err := sqlServerClient.Get(azConfig.ResourceGroup, deployConfig.ServerName)

	if err != nil {

		//glog.Warning(err)
		glog.Warning(err)
		de := err.(autorest.DetailedError)
		glog.Info(de.StatusCode)

		if de.StatusCode == http.StatusNotFound {

			glog.Info("Server not found, creating one")

			// Create Server
			serverConfg := sql.Server{ ServerProperties: &sql.ServerProperties{} }
			serverConfg.AdministratorLogin = &deployConfig.AdministratorLogin
			serverConfg.AdministratorLoginPassword = &deployConfig.AdministratorLoginPassword
			serverConfg.Kind = &deployConfig.Kind
			serverConfg.Name = &deployConfig.ServerName

			deref := serverConfg

			glog.Info("Start creation")
			result, err := sqlServerClient.CreateOrUpdate(azConfig.ResourceGroup, deployConfig.ServerName, deref)
			glog.Info("Completed creation")

			if err != nil {
				glog.Fatal(err)
			}

			fmt.Printf("Server created %v", result.FullyQualifiedDomainName)
			server = result
			fmt.Print(serverConfg)
		}

		if de.StatusCode == http.StatusUnauthorized || de.StatusCode == http.StatusForbidden {
			glog.Error("Credentials supplied aren't valid or able to perform action", err)
			panic("UnauthorizedError")
		}
	}

	glog.Info(server)

	fmt.Print(server.Name)

	// sqlDbClient := sql.NewDatabasesClient(c.SubscriptionID)
	// sqlDbClient.Authorizer = auth

	// sqlClient.CreateOrUpdate(c.ResourceGroup, serverName, databaseName)

	// sqlClient.Get(os.Getenv("AZURE_RESOURCE_GROUP"), )

	return "hello", nil
}

// NewServicePrincipalTokenFromCredentials creates a new ServicePrincipalToken using values of the
// passed credentials map.
func newServicePrincipalTokenFromCredentials(c azureProviders.ARMConfig, scope string) (*adal.ServicePrincipalToken, error) {
	oauthConfig, err := adal.NewOAuthConfig(azure.PublicCloud.ActiveDirectoryEndpoint, c.TenantID)
	if err != nil {
		panic(err)
	}
	return adal.NewServicePrincipalToken(*oauthConfig, c.ClientID, c.ClientSecret, scope)
}

var letters = []rune("abcdefghijklmnopqrstuvwxyz1234567890")

func randAlphaNumericSeq(n int) string {
	b := make([]rune, n)
	rand.Seed(time.Now().UnixNano())
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
