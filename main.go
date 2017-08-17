package main

import (
	"flag"
	"fmt"
	"time"

	"os"

	"github.com/golang/glog"
	"github.com/lawrencegripper/kube-azureresources/azureProviders"
	"github.com/lawrencegripper/kube-azureresources/crd"
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
		if  _, err := os.Stat(kubeconfig); !os.IsNotExist(err) {
			return clientcmd.BuildConfigFromFlags("", kubeconfig)
		}
	}
	
	return rest.InClusterConfig()
}

var providers map[string]azureProviders.Provider

func init() {
	// We log to stderr because glog will default to logging to a file.
	// By setting this debugging is easier via `kubectl logs`
	flag.Set("logtostderr", "true")
	flag.Set("stderrthreshold", "INFO")
	flag.Parse()
}

var clientConfig *rest.Config
var kubeClient *kubernetes.Clientset
var kubeCustomClient *rest.RESTClient

func main() {
	defer glog.Flush()

	exitChannel := make(chan int)

	fmt.Println("Load providers")

	providers = map[string]azureProviders.Provider{
		"cosmos": azureProviders.CosmosProvider{},
		"postgres": azureProviders.PostgresProvider{},
	}

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

	kubeClient = client

	if err != nil {
		glog.Fatalf("Failed to retreive nodes: %v", err)
	}

	kubeCustomClient, _, _ = crd.NewRestClient(clientConfig)

	azureResourceWatch := cache.NewListWatchFromClient(kubeCustomClient, "azureresources", api.NamespaceAll, fields.Everything())
	eStore, eController := cache.NewInformer(azureResourceWatch, &crd.AzureResource{}, time.Second*0, cache.ResourceEventHandlerFuncs{
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
	resourceUpdated(nil, a)
}

func resourceDeleted(a interface{}) {
	resource := a.(*crd.AzureResource)

	

	// .Println("Item deleted")
	fmt.Printf("Name: %v \n", resource.Name)
}

func resourceUpdated(oldItem, newItem interface{}) {
	resource := newItem.(*crd.AzureResource)
	
	fmt.Println("Item Updated")
	fmt.Printf("Name: %v \n", resource.Name)
	
	if resource.Status.ProvisioningStatus == "Provisioned" {
		//Todo: Handle and reconcile changes in both kube and azure to update resources.
		return
	}
	
	azCon, err := azureProviders.GetAzureConfigFromEnv()
	if err != nil {
		glog.Error(err)
		return
	}

	provider, exists := providers[resource.Spec.ResourceProvider]
	if !exists {
		glog.Error("Cannot progress resource of this type", resource)
		return
	}

	output, err := provider.CreateOrUpdate(azCon, *resource)

	if err != nil {
		glog.Error(err)
		return
	}

	k := azureProviders.NewKubeMan(clientConfig)
	k.Update(*resource, output, err)
}
