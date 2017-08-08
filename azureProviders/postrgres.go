package azureProviders

import (
	"github.com/lawrencegripper/kube-azureresources/crd"
	"net/http"

	"github.com/lawrencegripper/kube-azureresources/models"

	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/Azure/azure-sdk-for-go/arm/postgresql"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/golang/glog"
)

type PostgresConfig struct {
	ServerName                 string
	Location                   string
	AdministratorLogin         string
	AdministratorLoginPassword string
	Tags                       map[string]*string
}

func NewPostgresConfig(azRes crd.AzureResource) PostgresConfig {
	config := PostgresConfig{
		ServerName: randFromSelection(12, lettersLower),
		Location:   azRes.Spec.Location,
		Tags: azRes.GenerateAzureTags(),
		AdministratorLogin: "azurePostgres",
		AdministratorLoginPassword: randAlphaNumericSeq(24),
	}
	return config
}

func DeployPostgres(deployConfig PostgresConfig, azConfig ARMConfig) (models.Output, error) {
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
				CreateMode:                 "Default",
				SslEnforcement:             "Enabled",
				StorageMB:                  Int64Pointer(51200),
				Version:                    postgresql.NineFullStopSix}

			serverSku := postgresql.Sku{
				Capacity: Int32Pointer(50),
				Family:   StringPointer("SkuFamily"),
				Name:     StringPointer("PGSQLB50"),
				Tier:     "Basic",
				Size:     StringPointer("51200")}
			serverConfigCreate := postgresql.ServerForCreate{Sku: &serverSku, Location: &deployConfig.Location, Properties: serverPropertiesForCreate}
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
				case <-time.After(time.Minute * 6):
					glog.Error("Timeout occured creating server")
					return output, errors.New("Timout Occurred provisioning resource")
				}
			}

		}

		if de.StatusCode == http.StatusUnauthorized || de.StatusCode == http.StatusForbidden {
			glog.Error("Credentials supplied aren't valid or able to perform action", err)
			panic("UnauthorizedError")
		}
	}

	glog.Info(server)

	fmt.Print(server.Name)

	output = models.Output{
		Endpoint: *server.FullyQualifiedDomainName,
		Port:     5432,
		Secrets: map[string]string{
			"username": deployConfig.AdministratorLogin + "@" + deployConfig.ServerName,
			"password": deployConfig.AdministratorLoginPassword,
		},
		AzureResourceIds: []string{*server.ID},
	}

	return output, nil
}

//This is probably a pretty nasty hack, was interesting to play with runes.
//Todo: Simplify Password generation.
var lettersLower = []rune("abcdefghijklmnopqrstuvwxyz")
var lettersUpper = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
var numbers = []rune("1234567890")
var symbols = []rune("!@£%$£^&*_+")

func randAlphaNumericSeq(n int) string {
	bucketSize := n / 4
	return randFromSelection(bucketSize, lettersUpper) + randFromSelection(bucketSize, lettersLower) + randFromSelection(bucketSize, numbers) + randFromSelection(bucketSize, symbols)
}

func randFromSelection(length int, choices []rune) string {
	b := make([]rune, length)
	rand.Seed(time.Now().UnixNano())
	for i := range b {
		b[i] = choices[rand.Intn(len(choices))]
	}
	return string(b)
}
