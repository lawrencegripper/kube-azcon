package main

import (
	"flag"

	"fmt"

	"github.com/golang/glog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func GetClientConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}

func main() {
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

	nodes, err := client.Nodes().List(metav1.ListOptions{})

	if err != nil {
		glog.Fatalf("Failed to retreive nodes: %v", err)
	}

	fmt.Print(nodes.Items[0].Name)
}
