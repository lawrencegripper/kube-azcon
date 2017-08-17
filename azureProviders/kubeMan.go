package azureProviders

import (
	"errors"

	// "github.com/Azure/go-autorest/autorest"
	// "github.com/Azure/go-autorest/autorest/azure"

	"reflect"

	"github.com/lawrencegripper/kube-azureresources/crd"
	"github.com/lawrencegripper/kube-azureresources/models"
	"k8s.io/client-go/dynamic"

	"github.com/golang/glog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
)

type KubeMan struct {
	client        *kubernetes.Clientset
	config        *rest.Config
	resourceClient *dynamic.ResourceClient
}

//NewKubeMan - Creates a new kubeman for interfacing with the cluster.
func NewKubeMan(config *rest.Config) KubeMan {
	return KubeMan{
		client:        kubernetes.NewForConfigOrDie(config),
		resourceClient: crd.NewDynamicClientOrDie(config).Resource(&metav1.APIResource{
		Kind:       "AzureResource",
		Name:       "azureresources",
		Namespaced: true,
		}, "default"),
		config:        config,
	}
}

func (k *KubeMan) IsUptodate(azResource crd.AzureResource) (isUptodate *bool, err error) {

	return nil, errors.New("Not implimented yet")
}

func (k *KubeMan) Delete(azResource crd.AzureResource) (succeeded *bool, err error) {

	

	return nil, errors.New("Not implimented yet")
}

func (k *KubeMan) Update(azResource crd.AzureResource, serviceOutput models.Output, err error) {
	if err != nil {
		azResource.Status.ProvisioningStatus = "Error" + err.Error()
		resU, _ := azResource.AsUnstructured()
		k.resourceClient.Update(resU)

	} else {
		k.updateCrd(azResource, serviceOutput)
		k.addServiceAndSecrets(azResource, serviceOutput)
	}
}

func (k *KubeMan) updateCrd(azResource crd.AzureResource, serviceOutput models.Output) {
	resUnstructured, err := k.resourceClient.Get(azResource.Name)

	if err != nil {
		glog.Error(err)
		panic(err)
	}

	resource, err := crd.AzureResourceFromUnstructured(resUnstructured)

	if err != nil {
		panic(err)
	}

	resource.Status.ProvisioningStatus = "Provisioned"
	resource.Status.Output = serviceOutput

	// Investigate moving to patch command. Could concurrent read and update be overwritten currenctly?
	resToUpdate, _ := resource.AsUnstructured()

	//Only update if status has changed.
	if !reflect.DeepEqual(resource.Status, azResource.Status) {

		updateRes, err := k.resourceClient.Update(resToUpdate)
		if err != nil {
			glog.Error("Failed to update resource")
			glog.Error(err)
		} else {
			glog.Info("Updated azure resource")
			glog.Info(updateRes)
		}
	}

	glog.Info("Resource unchanged. Not updating in kube")
}

func (k *KubeMan) addServiceAndSecrets(azResource crd.AzureResource, serviceOutput models.Output) {
		
	service := v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: azResource.Name,
		},
		Spec: v1.ServiceSpec{
			Type:         "ExternalName",
			ExternalName: serviceOutput.Endpoint,
		},
	}

	srvRes, err := k.client.Services(azResource.Namespace).Create(&service)
	//Hack: If we can't create it try and udpate it. 
	if err != nil {
		//Get existing service and update it with endpoint. 
		srvCurrent, _ := k.client.Services(azResource.Namespace).Get(azResource.Name, metav1.GetOptions{})
		srvCurrent.Spec.ExternalName = serviceOutput.Endpoint
		srvRes, err = k.client.Services(azResource.Namespace).Update(srvCurrent)
	}

	if err != nil {
		glog.Error("Failed creating service")
		glog.Error(err)
	} else {
		glog.Info("Created Service")
		glog.Info(srvRes)
	}

	secret := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: azResource.Name,
		},
		Data: serviceOutput.GetSecretMap(),
	}

	secretRes, err := k.client.Secrets(azResource.Namespace).Create(&secret)
	//Hack: If we can't create it try and update it. 
	if err != nil {
		currentSecret, _ := k.client.Secrets(azResource.Namespace).Get(azResource.Name, metav1.GetOptions{})
		currentSecret.Data = serviceOutput.GetSecretMap()
		secretRes, err = k.client.Secrets(azResource.Namespace).Update(currentSecret)
	}

	if err != nil {
		glog.Error("Failed creating secret")
		glog.Error(err)
	} else {
		glog.Info("Created secret")
		glog.Info(secretRes)

	}

}
