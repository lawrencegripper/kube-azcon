package models

import (
	"encoding/base64"
)

type SecretsMap map[string]string

type Output struct {
	Secrets          SecretsMap `json:"secrets"`
	Port             int        `json:"port"`
	Endpoint         string     `json:"endpoint"`
	AzureResourceIds []string   `json:"azureResourceIds"`
}

func (s *Output) GetSecretMap() map[string][]byte {
	res := make(map[string][]byte)
	for k, v := range s.Secrets {
		bytesInString := []byte(v)
		res[k] = []byte(base64.StdEncoding.EncodeToString(bytesInString))
	}
	return res
}
