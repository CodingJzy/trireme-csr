package signer

import (
	"fmt"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	apiv1 "k8s.io/api/core/v1"

	certificatev1alpha1 "github.com/aporeto-inc/trireme-csr/apis/v1alpha1"
)

type certificateController struct {
	certificateClient *rest.RESTClient
}

func (c *certificateController) Run() error {
	fmt.Print("Watch Certificates objects\n")

	// Watch Example objects
	_, err := c.watchCerts()
	if err != nil {
		fmt.Printf("Failed to register watch for Example resource: %v\n", err)
		return err
	}

	return nil
}

func (c *certificateController) watchCerts() (cache.Controller, error) {
	source := cache.NewListWatchFromClient(
		c.certificateClient,
		certificatev1alpha1.CertificateResourcePlural,
		apiv1.NamespaceAll,
		fields.Everything())

	_, controller := cache.NewInformer(
		source,

		// The object type.
		&certificatev1alpha1.Certificate{},

		0,

		cache.ResourceEventHandlerFuncs{
			AddFunc:    c.onAdd,
			UpdateFunc: c.onUpdate,
			DeleteFunc: c.onDelete,
		})

	go controller.Run(nil)
	return controller, nil
}

func (c *certificateController) onAdd(obj interface{}) {
	fmt.Printf("AddingCert: %v\n", obj)

}

func (c *certificateController) onUpdate(oldObj, newObj interface{}) {
	fmt.Printf("UpdatingCert: %v\n", newObj)
}

func (c *certificateController) onDelete(obj interface{}) {
	fmt.Printf("DeletingCert: %v\n", obj)
}
