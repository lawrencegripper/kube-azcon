package postgresProvider

import (
	"net/http"

	"fmt"
	"math/rand"
	"time"

	"github.com/Azure/azure-sdk-for-go/arm/postgresql"
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
	config.AdministratorLoginPassword = randAlphaNumericSeq(24)
	return config
}

func StringPointer(i string) *string { return &i }
func Int32Pointer(i int32) *int32    { return &i }
func IntPointer(i int) *int          { return &i }

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
				StorageMB:                  func(i int64) *int64 { return &i }(50),
				Version:                    postgresql.NineFullStopSix}

			serverSku := postgresql.Sku{
				Capacity: Int32Pointer(50),
				Family:   StringPointer("SkuFamily"),
				Name:     StringPointer("PGSQLB50"),
				Tier:     "Basic",
				Size:     StringPointer("50")}
			serverConfigCreate := postgresql.ServerForCreate{Sku: &serverSku, Location: &deployConfig.Location, Properties: serverPropertiesForCreate}
			//detailsType := "Microsoft.DBforPostgreSQL/servers"
			//serverConfg.Type = &detailsType

			glog.Info("Start creation")
			cancelChannel := make(chan struct{})
			resultChan, errChan := sqlServerClient.CreateOrUpdate(
				azConfig.ResourceGroup, deployConfig.ServerName, serverConfigCreate, cancelChannel)
			glog.Info("Completed creation")

			//fmt.Printf("Server created %v", result)
			for index := 0; index < 2; index++ {
				select {
				case err := <-errChan:
					glog.Info(err)
				case res := <-resultChan:
					glog.Info(res.Name)
				case <-time.After(time.Second * 360):
					glog.Info("Timeout occured creating server")
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
