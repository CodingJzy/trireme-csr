package controller

import (
	"fmt"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"

	"go.uber.org/zap"

	certificatev1alpha1 "github.com/aporeto-inc/trireme-csr/apis/v1alpha1"
	"github.com/aporeto-inc/trireme-csr/certificates"
	certificateclient "github.com/aporeto-inc/trireme-csr/client"

	"github.com/aporeto-inc/tg/tglib"
)

// CertificateController contains all the logic to implement the issuance of certificates.
type CertificateController struct {
	certificateClient *certificateclient.CertificateClient
	issuer            certificates.Issuer
}

var certPath = "/Users/bvandewa/golang/src/github.com/aporeto-inc/trireme-csr/testdata/private/ca.cert.pem"
var certKeyPath = "/Users/bvandewa/golang/src/github.com/aporeto-inc/trireme-csr/testdata/private/ca.key.pem"
var certPass = "test"

// NewCertificateController generates the new CertificateController
func NewCertificateController(certificateClient *certificateclient.CertificateClient, issuer certificates.Issuer) (*CertificateController, error) {

	return &CertificateController{
		certificateClient: certificateClient,
		issuer:            issuer,
	}, nil
}

// Run starts the certificateWatcher.
func (c *CertificateController) Run() error {
	zap.L().Info("start watching Certificates objects")

	// Watch Certificate objects
	_, err := c.watchCerts()
	if err != nil {
		return fmt.Errorf("failed to register watch for Certificate CRD resource: %s", err)
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
	csr, err := tglib.LoadCSR(certRequest.Spec.Request)
	if err != nil {
		zap.L().Error("Error loading CSR", zap.Error(err), zap.String("namespace", certRequest.Namespace), zap.String("name", certRequest.Name))
		return
	}

	zap.L().Info("Validating cert request", zap.String("namespace", certRequest.Namespace), zap.String("name", certRequest.Name))

	err = c.issuer.Validate(csr)
	if err != nil {
		zap.L().Error("CSR has not been validated", zap.Error(err), zap.String("namespace", certRequest.Namespace), zap.String("name", certRequest.Name))
		return
	}
	zap.L().Info("Cert request has been accepted", zap.String("namespace", certRequest.Namespace), zap.String("name", certRequest.Name))

	cert, err := c.issuer.Sign(csr)
	if err != nil {
		zap.L().Error("Error loading CSR", zap.Error(err), zap.String("namespace", certRequest.Namespace), zap.String("name", certRequest.Name))
		return
	}
	zap.L().Info("Cert successfully generated", zap.String("namespace", certRequest.Namespace), zap.String("name", certRequest.Name))

	zap.L().Debug("Cert successfully generated", zap.ByteString("cert", cert))

	certRequest.Status.Certificate = cert
	certRequest.Status.Ca = c.issuer.GetCACert()
	certRequest.Status.State = certificatev1alpha1.CertificateStateCreated

	c.certificateClient.Certificates(certRequest.Namespace).Update(certRequest)
}

func (c *CertificateController) onUpdate(oldObj, newObj interface{}) {
	zap.L().Debug("UpdatingCert")
	certRequest := newObj.(*certificatev1alpha1.Certificate)

	// Checking if the Status is already a generated Cert:
	if certRequest.Status.Certificate != nil {
		zap.L().Debug("Updated Cert request has already been processed", zap.String("namespace", certRequest.Namespace), zap.String("name", certRequest.Name))
	}

	zap.L().Info("Cert Request updated still has to be generated", zap.String("namespace", certRequest.Namespace), zap.String("name", certRequest.Name))

}

func (c *CertificateController) onDelete(obj interface{}) {
	zap.L().Debug("DeletingCert")
}
