package azureProviders

import (
	"strings"
	"net/http"

	"github.com/lawrencegripper/kube-azcon/crd"

	"github.com/lawrencegripper/kube-azcon/models"

	"errors"
	"time"

	"github.com/Azure/azure-sdk-for-go/arm/postgresql"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/golang/glog"
)

type PostgresProvider struct { }

func (p PostgresProvider) CreateOrUpdate(azConfig ARMConfig, azRes crd.AzureResource) (models.Output, error) {
	deployConfig := newPostgresConfig(azConfig, azRes)
	return deployPostgres(deployConfig, azConfig)
}


type postgresConfig struct {
	ServerName                 string
	Location                   string
	AdministratorLogin         string
	AdministratorLoginPassword string
	Tags                       map[string]*string
}



func newPostgresConfig(azConfig ARMConfig, azRes crd.AzureResource) postgresConfig {
	config := postgresConfig{
		ServerName:                 strings.ToLower(azConfig.ResourcePrefix + azRes.Name),
		Location:                   azRes.Spec.Location,
		Tags:                       azRes.GenerateAzureTags(),
		AdministratorLogin:         "azurePostgres",
		AdministratorLoginPassword: randAlphaNumericSeq(16),
	}
	return config
}

func deployPostgres(deployConfig postgresConfig, azConfig ARMConfig) (models.Output, error) {
	var output models.Output

	azConfig, err := GetAzureConfigFromEnv()

	if err != nil {
		glog.Fatalf("Failed to get azure configuration from environment: %v", err)
	}

	spt, err := NewServicePrincipalTokenFromCredentials(azConfig, azure.PublicCloud.ResourceManagerEndpoint)

	if err != nil {
		glog.Fatalf("Failed creating service principal: %v", err)
	}

	auth := autorest.NewBearerAuthorizer(spt)

	sqlServerClient := postgresql.NewServersClient(azConfig.SubscriptionID)
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
			serverPropertiesForCreate := postgresql.ServerPropertiesForDefaultCreate{
				AdministratorLogin:         &deployConfig.AdministratorLogin,
				AdministratorLoginPassword: &deployConfig.AdministratorLoginPassword,
				StorageMB:                  Int64Pointer(51200),
				Version:                    postgresql.NineFullStopSix}

			serverSku := postgresql.Sku{
				Capacity: Int32Pointer(50),
				Family:   StringPointer("SkuFamily"),
				Name:     StringPointer("PGSQLB50"),
				Tier:     postgresql.Basic,
				Size:     StringPointer("51200")}

			serverConfigCreate := postgresql.ServerForCreate{Sku: &serverSku, Location: &deployConfig.Location, Properties: &serverPropertiesForCreate}
			serverConfigCreate.Tags = &deployConfig.Tags
			//detailsType := "Microsoft.DBforPostgreSQL/servers"
			//serverConfg.Type = &detailsType

			glog.Info("Start creation")
			cancelChannel := make(chan struct{})
			resultChan, errChan := sqlServerClient.CreateOrUpdate(
				azConfig.ResourceGroup, deployConfig.ServerName, serverConfigCreate, cancelChannel)

			//Refactor this not sure it's necessary.
			//Think it can be done better
			for index := 0; index < 2; index++ {
				select {
				case err := <-errChan:
					glog.Error(err)
					return output, err
				case res := <-resultChan:
					server = res
					glog.Info("Completed creation")
					break
				case <-time.After(time.Minute * 12):
					glog.Error("Timeout occured creating server")
					return output, errors.New("Timout Occurred provisioning resource")
				}
			}

		}

		if de.StatusCode == http.StatusUnauthorized || de.StatusCode == http.StatusForbidden {
			glog.Error("Credentials supplied aren't valid or able to perform action", err)
			panic("UnauthorizedError")
		}
	} else {
		glog.Info("Server already exists")
	}
	

	output = models.Output{
		Endpoint: *server.FullyQualifiedDomainName,
		Port:     5432,
		Secrets: map[string]string{
			"username": deployConfig.AdministratorLogin + "@" + deployConfig.ServerName,
			"password": deployConfig.AdministratorLoginPassword,
			"endpoint": *server.FullyQualifiedDomainName,
			"port":     "5432",
		},
		AzureResourceIds: []string{*server.ID},
	}

	return output, nil
}
