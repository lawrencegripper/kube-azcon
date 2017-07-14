package integration

import (
	"testing"
	"github.com/lawrencegripper/kube-azureresources/azureProviders/sql"
)

func TestSqlProviderCreatesResource(t *testing.T) {
	t.Parallel()

	result, err := sql.Deploy("testServer","testDb","westEurope")

	if err != nil {
		t.Error("Failed creating sql deployment")
	}

	t.Log(result);
	
}