package azureProviders

import (
	"github.com/lawrencegripper/kube-azureresources/crd"

	"github.com/lawrencegripper/kube-azureresources/models"

	"errors"
	"time"

	"github.com/Azure/azure-sdk-for-go/arm/cosmos-db"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/golang/glog"
)

type CosmosConfig struct {
	AccountName string
	Location    string
	Tags        map[string]*string
	KubeLink    string
}

func NewCosmosConfig(azConfig ARMConfig, azRes crd.AzureResource, ) CosmosConfig {
	config := CosmosConfig{
		AccountName: azConfig.ResourcePrefix + azRes.Name,
		Location:    azRes.Spec.Location,
		Tags:        azRes.GenerateAzureTags(),
		KubeLink:    azRes.SelfLink,
	}

	// temp add default experince
	config.Tags["defaultExperience"] = StringPointer("MongoDB")
	return config
}

func DeployCosmos(deployConfig CosmosConfig, azConfig ARMConfig) (models.Output, error) {
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

	client := cosmosdb.NewDatabaseAccountsClient(azConfig.SubscriptionID)
	client.Authorizer = auth

	accounts, err := client.ListByResourceGroup(azConfig.ResourceGroup)

	if err != nil {
		return output, errors.New("Failed to enumerate accounts")
	}

	var dbAccount cosmosdb.DatabaseAccount
	accountExists := false
	// check for nil pointer (todo: double check this is the best way to do this)

	// find the right cosmos account, if it exists
	if accounts.Value != nil {
		for _, v := range *accounts.Value {
			if v.Tags == nil {
				continue
			}
			tags := *v.Tags
			kubeLink, exists := tags[crd.TagKubernetesResourceLink]
			if exists && *kubeLink == deployConfig.KubeLink {
				dbAccount = v
				accountExists = true
				break
			}
		}
	}

	glog.Info("Cosmos Account Exists", accountExists, deployConfig.AccountName)

	if !accountExists {

		cancelChannel := make(chan struct{})
		locationID := deployConfig.AccountName + deployConfig.Location
		properties := cosmosdb.DatabaseAccountCreateUpdateParameters{
			Location: &deployConfig.Location,
			Tags:     &deployConfig.Tags,
			Kind:     cosmosdb.MongoDB,
		}
		properties.DatabaseAccountCreateUpdateProperties = &cosmosdb.DatabaseAccountCreateUpdateProperties{
			DatabaseAccountOfferType: StringPointer("Standard"),
			Locations: &[]cosmosdb.Location{
				cosmosdb.Location{
					LocationName: &deployConfig.Location,
					ID:           &locationID,
				},
			},
		}

		glog.Info("Starting Cosmos Deployment")
		resultChan, errChan := client.CreateOrUpdate(azConfig.ResourceGroup, deployConfig.AccountName, properties, cancelChannel)

		//Refactor this not sure it's necessary.
		//Think it can be done better
		for index := 0; index < 2; index++ {
			select {
			case err := <-errChan:
				if err == nil {
					continue
				}
				glog.Error(err)
				return output, err
			case res := <-resultChan:
				dbAccount = res
				glog.Info("Completed creation")
			case <-time.After(time.Minute * 10):
				glog.Error("Timeout occured creating server")
				return output, errors.New("Timout Occurred provisioning resource")
			}
		}

	} else {
		switch *dbAccount.ProvisioningState{
			case "Succeeded":
				glog.Info("Resource already deployed. Status:", *dbAccount.ProvisioningState)
			default:
				glog.Error("Todo: Handle other deployment states for the resource", *dbAccount.ProvisioningState)
		}
	}


	glog.Info("Getting connection strings")
	connectionStrings, err := client.ListConnectionStrings(azConfig.ResourceGroup, *dbAccount.Name)

	if err != nil {
		return output, errors.New("Failed to list connection strings for dbaccount:" + *dbAccount.Name)
	}

	glog.Info("Connection strings")

	secrets := map[string]string{}
	if connectionStrings.ConnectionStrings != nil{
		for _, con := range *connectionStrings.ConnectionStrings {
			secrets[*con.Description] = *con.ConnectionString
		}
	}



	output = models.Output{
		Endpoint:         *dbAccount.DocumentEndpoint,
		Port:             5432,
		Secrets:          secrets,
		AzureResourceIds: []string{*dbAccount.ID},
	}

	return output, nil
}
