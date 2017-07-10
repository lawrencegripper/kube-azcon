package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/golang/glog"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"github.com/lawrencegripper/kube-azureresources/crd"
)

func GetClientConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}

func main() {
	fmt.Println("hello world")

	// When running as a pod in-cluster, a kubeconfig is not needed. Instead this will make use of the service account injected into the pod.
	// However, allow the use of a local kubeconfig as this can make local development & testing easier.
	kubeconfig := flag.String("kubeconfig", "/Users/lawrence/.kube/config.d/sharedcluster.json", "Path to a kubeconfig file")

	// We log to stderr because glog will default to logging to a file.
	// By setting this debugging is easier via `kubectl logs`
	flag.Set("logtostderr", "true")
	flag.Parse()

	// nodeNameEnv := "Barry"
	// // The node name is necessary so we can identify "self".
	// // This environment variable is assumed to be set via the pod downward-api, however it can be manually set during testing
	// nodeName := os.Getenv(nodeNameEnv)
	// if nodeName == "" {
	// 	glog.Fatalf("Missing required environment variable %s", nodeNameEnv)
	// }

	// Build the client config - optionally using a provided kubeconfig file.

	clientConfig, err := GetClientConfig(*kubeconfig)
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

	clientCustom, _, _ := crd.NewClient(clientConfig)

	azureResourceWatch := cache.NewListWatchFromClient(clientCustom, "azureresources", api.NamespaceAll, fields.Everything())
	resyncPeriod := 10 * time.Second
	eStore, eController := cache.NewInformer(azureResourceWatch, &crd.AzureResource{}, resyncPeriod, cache.ResourceEventHandlerFuncs{
		AddFunc:    resourceCreated,
		DeleteFunc: resourceDeleted,
	})

	//Run the controller as a goroutine
	go eController.Run(wait.NeverStop)

	for !eController.HasSynced() {
		fmt.Println("Waiting for sync")
		time.Sleep(15 * time.Second)
	}

	resources := eStore.List()
	for index := 0; index < len(resources); index++ {
		obj := resources[index]
		resource := obj.(*crd.AzureResource)
		fmt.Println(resource.Name)
	}
}

func resourceCreated(a interface{}) {

}

func resourceDeleted(a interface{}) {

}
