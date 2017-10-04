package controller

import (
	"fmt"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"

	certificatev1alpha1 "github.com/aporeto-inc/trireme-csr/apis/v1alpha1"
	"github.com/aporeto-inc/trireme-csr/certificates"
	certificateclient "github.com/aporeto-inc/trireme-csr/client"
)

// CertificateController contains all the logic to implement the issuance of certificates.
type CertificateController struct {
	certificateClient *certificateclient.CertificateClient
	issuer            *certificates.Issuer
}

var certPath = "/Users/bvandewa/golang/src/github.com/aporeto-inc/trireme-csr/testdata/private/ca.cert.pem"
var certKeyPath = "/Users/bvandewa/golang/src/github.com/aporeto-inc/trireme-csr/testdata/private/ca.key.pem"
var certPass = "test"

// NewCertificateController generates the new CertificateController
func NewCertificateController(certificateClient *certificateclient.CertificateClient, ca string) *CertificateController {

	issuer, err := certificates.NewIssuerFromPath(certPath, certKeyPath, certPass)
	if err != nil {
		fmt.Printf("Error creating new Issuer %s", err)
	}

	return &CertificateController{
		certificateClient: certificateClient,
		issuer:            issuer,
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
		c.certificateClient.RESTClient(),
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
	csr, err := certificates.LoadCSR(certRequest.Spec.Request)
	if err != nil {
		fmt.Printf("Error loading cert: %s\n", err)
	}

	cert, err := c.issuer.Sign(csr)
	if err != nil {
		fmt.Printf("Error issuing cert: %s\n", err)
	}
	fmt.Printf("Cert generated: %s\n", cert)

	certRequest.Status.Certificate = cert
	certRequest.Status.State = certificatev1alpha1.CertificateStateCreated

	c.certificateClient.Certificates(certRequest.Namespace).Update(certRequest)
}

func (c *CertificateController) onUpdate(oldObj, newObj interface{}) {
	fmt.Printf("UpdatingCert: %v\n", newObj)
}

func (c *CertificateController) onDelete(obj interface{}) {
	fmt.Printf("DeletingCert: %v\n", obj)
}
