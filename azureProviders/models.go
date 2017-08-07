package azureProviders

type SecretsMap map[string]string

type Output struct {
	Secrets SecretsMap
	Port int
	Endpoint string
}

