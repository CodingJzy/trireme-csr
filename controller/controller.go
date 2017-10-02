package controller

import (
	"fmt"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	certificatev1alpha1 "github.com/aporeto-inc/trireme-csr/apis/v1alpha1"
)

type CertificateController struct {
	certificateClient rest.Interface
}

// NewCertificateController generates the new CertificateController
func NewCertificateController(certificateClient rest.Interface, ca string) *CertificateController {
	return &CertificateController{
		certificateClient: certificateClient,
	}
}

// Run starts the certificateWatcher.
func (c *CertificateController) Run() error {
	fmt.Print("Watch Certificates objects\n")

	// Watch Example objects
	_, err := c.watchCerts()
	if err != nil {
		fmt.Printf("Failed to register watch for Example resource: %v\n", err)
		return err
	}

	return nil
}

func (c *CertificateController) watchCerts() (cache.Controller, error) {
	source := cache.NewListWatchFromClient(
		c.certificateClient,
		certificatev1alpha1.CertificateResourcePlural,
		"",
		fields.Everything())

	_, controller := cache.NewInformer(
		source,

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

func (c *CertificateController) onAdd(obj interface{}) {
	fmt.Printf("AddingCert: %v\n", obj)
	certRequest := obj.(*certificatev1alpha1.Certificate)
	fmt.Printf("AddingCert: %v\n", certRequest)
}

func (c *CertificateController) onUpdate(oldObj, newObj interface{}) {
	fmt.Printf("UpdatingCert: %v\n", newObj)
}

func (c *CertificateController) onDelete(obj interface{}) {
	fmt.Printf("DeletingCert: %v\n", obj)
}
