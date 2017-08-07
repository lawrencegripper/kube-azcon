package crd

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

var versionedGroupName = schema.GroupVersion{Group: "stable.gripdev.xyz", Version: "v1"}

func NewRestClient(cfg *rest.Config) (*rest.RESTClient, *runtime.Scheme, error) {

	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(versionedGroupName, &AzureResource{}, &AzureResourceList{})

	config := *cfg
	config.GroupVersion = &versionedGroupName
	config.APIPath = "/apis"
	config.ContentType = runtime.ContentTypeJSON
	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: serializer.NewCodecFactory(scheme)}

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, nil, err
	}

	return client, scheme, nil
}

func NewDynamicClientOrDie(cfg *rest.Config) (*dynamic.Client) {

	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(versionedGroupName, &AzureResource{}, &AzureResourceList{})

	config := *cfg
	config.GroupVersion = &versionedGroupName
	config.APIPath = "/apis"
	config.ContentType = runtime.ContentTypeJSON
	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: serializer.NewCodecFactory(scheme)}

	client, err := dynamic.NewClient(&config)
	if err != nil {
		panic(err)
	}

	return client
}
