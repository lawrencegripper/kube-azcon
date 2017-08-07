package azureProviders

import (
	"k8s.io/client-go/dynamic"
	"time"
	"github.com/lawrencegripper/kube-azureresources/crd"
	"github.com/lawrencegripper/kube-azureresources/models"

	"github.com/golang/glog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
)

type KubeMan struct {
	client *kubernetes.Clientset
	config *rest.Config
	dynamicClient *dynamic.Client
}

//NewKubeMan - Creates a new kubeman for interfacing with the cluster.
func NewKubeMan(config *rest.Config) KubeMan {
	return KubeMan{
		client: kubernetes.NewForConfigOrDie(config),
		dynamicClient: crd.NewDynamicClientOrDie(config),
		config: config,
	}
}

func (k *KubeMan) Update(resourceName string, serviceOutput models.Output){
	k.updateCrd(resourceName, serviceOutput)
	k.addServiceAndSecrets(serviceOutput)
}

func (k *KubeMan) updateCrd(resourceName string, serviceOutput models.Output) {

	azResourceClient := k.dynamicClient.Resource(&metav1.APIResource{
		Kind:       "AzureResource",
		Name:       "azureresources",
		Namespaced: true,
	}, "default")

	resUnstructured, err := azResourceClient.Get(resourceName)

	if err != nil{
		glog.Error(err)
		panic(err)
	}
	
	resource, err := crd.AzureResourceFromUnstructured(resUnstructured)

	if err != nil{
		panic(err)
	}

	resource.Status.ProvisioningStatus = "Provisioned"
	resource.Status.Output = serviceOutput
	resource.Status.LastChecked = time.Now()

	resToUpdate, err := resource.AsUnstructured()
	updateRes, err := azResourceClient.Update(resToUpdate)

	glog.Info("Updated azure resource")
	glog.Info(updateRes)
}

func (k *KubeMan) addServiceAndSecrets(serviceOutput models.Output) {
	service := v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceOutput.ServiceName,
		},
		Spec: v1.ServiceSpec{
			Type:         "ExternalName",
			ExternalName: serviceOutput.Endpoint,
		},
	}

	srvRes, err := k.client.Services(serviceOutput.Namespace).Create(&service)

	if err != nil {
		glog.Error("Failed creating service")
		glog.Error(err)
	} else {
		glog.Info("Created Service")
		glog.Info(srvRes)
	}

	secret := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceOutput.ServiceName,
		},
		Data: serviceOutput.GetSecretMap(),
	}

	secretRes, err := k.client.Secrets(serviceOutput.Namespace).Create(&secret)

	if err != nil {
		glog.Error("Failed creating secret")
		glog.Error(err)
	} else {
		glog.Info("Created secret")
		glog.Info(secretRes)

	}

}
