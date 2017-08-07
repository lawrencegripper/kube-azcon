package main

import (
	"flag"
	"fmt"
	"time"

	"os"

	"github.com/golang/glog"
	"github.com/lawrencegripper/kube-azureresources/azureProviders"
	"github.com/lawrencegripper/kube-azureresources/azureProviders/postgresProvider"
	"github.com/lawrencegripper/kube-azureresources/crd"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

func getClientConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}

func init() {
	// We log to stderr because glog will default to logging to a file.
	// By setting this debugging is easier via `kubectl logs`
	flag.Set("logtostderr", "true")
	flag.Set("stderrthreshold", "INFO")
	flag.Parse()
}

var kubeCustomClient *rest.RESTClient
var clientConfig *rest.Config

func main() {
	defer glog.Flush()

	exitChannel := make(chan int)

	fmt.Println("hello world")

	// When running as a pod in-cluster, a kubeconfig is not needed. Instead this will make use of the service account injected into the pod.
	// However, allow the use of a local kubeconfig as this can make local development & testing easier.
	kubeconfig := flag.String("kubeconfig", "/Users/lawrence/.kube/lgkube1", "Path to a kubeconfig file")

	glog.Info("Starting up....")

	var err error
	clientConfig, err = getClientConfig(*kubeconfig)
	if err != nil {
		glog.Fatalf("Failed to load client config: %v", err)
	}

	// Construct the Kubernetes client
	client, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		glog.Fatalf("Failed to create kubernetes client: %v", err)
	}

	nodes, err := client.Nodes().List(meta_v1.ListOptions{})

	if err != nil {
		glog.Fatalf("Failed to retreive nodes: %v", err)
	}

	fmt.Println("Nodes:")
	for index := 0; index < len(nodes.Items); index++ {
		fmt.Println(nodes.Items[index].Name)
	}

	kubeCustomClient, _, _ = crd.NewRestClient(clientConfig)

	azureResourceWatch := cache.NewListWatchFromClient(kubeCustomClient, "azureresources", api.NamespaceAll, fields.Everything())
	resyncPeriod := 10 * time.Second
	eStore, eController := cache.NewInformer(azureResourceWatch, &crd.AzureResource{}, resyncPeriod, cache.ResourceEventHandlerFuncs{
		AddFunc:    resourceCreated,
		DeleteFunc: resourceDeleted,
		UpdateFunc: resourceUpdated,
	})

	//Run the controller as a goroutine
	go eController.Run(wait.NeverStop)

	for !eController.HasSynced() {
		fmt.Println("Waiting for sync")
		time.Sleep(3 * time.Second)
	}

	resources := eStore.List()
	for index := 0; index < len(resources); index++ {
		obj := resources[index]
		resource := obj.(*crd.AzureResource)
		fmt.Println(resource.Name)
	}

	exitCode := <-exitChannel
	os.Exit(exitCode)
}

func resourceCreated(a interface{}) {
	resource := a.(*crd.AzureResource)

	azCon, err := azureProviders.GetConfigFromEnv()

	if err != nil {
		glog.Fatal(err)
	}

	if resource.Status.ProvisioningStatus != "Provisioned" {
		depCon := postgresProvider.NewPostgresConfig(azCon.ResourcePrefix+"testserver", azCon.ResourcePrefix+"testdb", "westeurope")

		result, err := postgresProvider.Deploy(depCon, azCon)

		client, _, err := crd.NewDynamicClient(clientConfig)

		azResourceClient := client.Resource(&meta_v1.APIResource{
			Kind:       "AzureResource",
			Name:       "azureresources",
			Namespaced: true,
		}, "default")

		resource.Status.ProvisioningStatus = "Provisioned"
		resource.Status.Output = result
		resource.Status.LastChecked = time.Now()

		resUnstructured, err := resource.AsUnstructured()
		updateRes, err := azResourceClient.Update(resUnstructured)

		glog.Info(updateRes)
		glog.Error(err)
	}
}

func resourceDeleted(a interface{}) {
	resource := a.(*crd.AzureResource)

	// if resource.Spec.ResourceProvider == "azuresql" {
	// 	result, err := sql.Deploy("","","")
	// }

	fmt.Println("Item deleted")
	fmt.Printf("Name: %v \n", resource.Name)
}

func resourceUpdated(oldItem, newItem interface{}) {
	resource := oldItem.(*crd.AzureResource)

	fmt.Println("Item Updated")
	fmt.Printf("Name: %v \n", resource.Name)
}
