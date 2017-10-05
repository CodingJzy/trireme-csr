package controller

import (
	"fmt"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"

	"go.uber.org/zap"

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
func NewCertificateController(certificateClient *certificateclient.CertificateClient, ca string) (*CertificateController, error) {

	issuer, err := certificates.NewIssuerFromPath(certPath, certKeyPath, certPass)
	if err != nil {
		return nil, fmt.Errorf("Error creating new Issuer %s", err)
	}

	return &CertificateController{
		certificateClient: certificateClient,
		issuer:            issuer,
	}, nil
}

// Run starts the certificateWatcher.
func (c *CertificateController) Run() error {
	zap.L().Warn("start watching Certificates objects")

	// Watch Example objects
	_, err := c.watchCerts()
	if err != nil {
		zap.L().Error("Failed to register watch for Certificate CRD resource: %v\n", zap.Error(err))
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
	zap.L().Debug("Adding Cert event")
	certRequest := obj.(*certificatev1alpha1.Certificate)
	csr, err := certificates.LoadCSR(certRequest.Spec.Request)
	if err != nil {
		zap.L().Error("Error loading CSR", zap.Error(err), zap.String("namespace", certRequest.Namespace), zap.String("name", certRequest.Name))
		return
	}

	cert, err := c.issuer.Sign(csr)
	if err != nil {
		zap.L().Error("Error loading CSR", zap.Error(err), zap.String("namespace", certRequest.Namespace), zap.String("name", certRequest.Name))
		return
	}
	zap.L().Info("Cert successfully generated", zap.ByteString("cert", cert))

	certRequest.Status.Certificate = cert
	certRequest.Status.State = certificatev1alpha1.CertificateStateCreated

	c.certificateClient.Certificates(certRequest.Namespace).Update(certRequest)
}

func (c *CertificateController) onUpdate(oldObj, newObj interface{}) {
	zap.L().Debug("UpdatingCert")
}

func (c *CertificateController) onDelete(obj interface{}) {
	zap.L().Debug("DeletingCert")
}
