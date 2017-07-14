package integration

import (
	"testing"

	"os"

	"github.com/golang/glog"
	"github.com/lawrencegripper/kube-azureresources/azureProviders"
	"github.com/lawrencegripper/kube-azureresources/azureProviders/sql"
	"flag"
)

func init() {
	// We log to stderr because glog will default to logging to a file.
	// By setting this debugging is easier via `kubectl logs`
	flag.Set("logtostderr", "true")
	flag.Set("stderrthreshold", "INFO")
	flag.Parse()
}

var resourcePrefix = os.Getenv("TEST_RESOURCE_PREFIX")

func TestSqlProviderCreatesResource(t *testing.T) {
	t.Parallel()

	glog.Info("Starting test...")
	azCon, err := azureProviders.GetConfigFromEnv()

	if err != nil {
		t.Error(err)
	}

	depCon := sql.NewPostgresConfig(resourcePrefix+"testserver", resourcePrefix+"testdb", "westeurope")

	result, err := sql.Deploy(depCon, azCon)

	if err != nil {
		t.Error("Failed creating sql deployment")
	}

	t.Log(result)



}
