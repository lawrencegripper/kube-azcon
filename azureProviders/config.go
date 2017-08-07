package azureProviders

import (
	"encoding/json"
	"os"
)

// ARMConfig - Basic azure config used to interact with ARM resources.
type ARMConfig struct {
	ClientID       string
	ClientSecret   string
	SubscriptionID string
	TenantID       string
	ResourceGroup  string
	ResourcePrefix string
}

// ConfigError - Error when reading configuration values.
type ConfigError struct {
	CurrentConfig ARMConfig
	ErrorDetails  string
}

func (e *ConfigError) Error() string {
	configJSON, err := json.Marshal(e.CurrentConfig)
	if err != nil {
		return e.ErrorDetails
	}

	return e.ErrorDetails + ": " + string(configJSON)
}

// GetConfigFromEnv - Retreives the azure configuration from environment variables.
func GetConfigFromEnv() (ARMConfig, error) {
	config := ARMConfig{
		ClientID:       os.Getenv("AZURE_CLIENT_ID"),
		ClientSecret:   os.Getenv("AZURE_CLIENT_SECRET"),
		ResourceGroup:  os.Getenv("AZURE_RESOURCE_GROUP"),
		SubscriptionID: os.Getenv("AZURE_SUBSCRIPTION_ID"),
		TenantID:       os.Getenv("AZURE_TENANT_ID"),
		ResourcePrefix: os.Getenv("TEST_RESOURCE_PREFIX"),
	}

	if config.ClientID == "" ||
		config.ClientSecret == "" ||
		config.ResourceGroup == "" ||
		config.SubscriptionID == "" ||
		config.TenantID == "" {
		return config, &ConfigError{CurrentConfig: config, ErrorDetails: "Missing configuration"}
	}

	return config, nil
}
