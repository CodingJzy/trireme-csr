package controller

import (
	"fmt"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"

	"go.uber.org/zap"

	certificatev1alpha2 "github.com/aporeto-inc/trireme-csr/apis/v1alpha2"
	"github.com/aporeto-inc/trireme-csr/certificates"
	certificateclient "github.com/aporeto-inc/trireme-csr/client"

	"github.com/aporeto-inc/tg/tglib"
)

// CertificateController contains all the logic to implement the issuance of certificates.
type CertificateController struct {
	certificateClient *certificateclient.CertificateClient
	issuer            certificates.Issuer
}

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
		certificatev1alpha2.CertificateResourcePlural,
		"",
		fields.Everything())

	_, controller := cache.NewInformer(
		source,
		&certificatev1alpha2.Certificate{},
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
	certRequest, ok := obj.(*certificatev1alpha2.Certificate)
	if !ok {
		zap.L().Sugar().Errorf("Received wrong object type in adding Cert event: '%T", obj)
		return
	}

	// if a new object is added which already has the signed or rejected phase, we don't need to take any actions
	// TODO: if a cert is added as signed, we should validate it before we allow this
	if certRequest.Status.Phase == certificatev1alpha2.CertificatePhaseSigned || certRequest.Status.Phase == certificatev1alpha2.CertificatePhaseRejected {
		zap.L().Debug("Added Cert request has already been processed", zap.String("name", certRequest.Name))
		return
	}

	// if a new object is added with phase "submitted", it means that we want to submit it for signing
	if certRequest.Status.Phase == certificatev1alpha2.CertificatePhaseSubmitted {
		zap.L().Info("Processing Cert request")
		c.processCert(certRequest)
		return
	}

	// otherwise nothing has to be done
	zap.L().Debug("Nothing has to be done for this Cert request")
}

func (c *CertificateController) onUpdate(oldObj, newObj interface{}) {
	zap.L().Debug("Updating Cert event")
	certRequest, ok := newObj.(*certificatev1alpha2.Certificate)
	if !ok {
		zap.L().Sugar().Errorf("Received wrong object type in updating Cert event: '%T", newObj)
		return
	}

	// Checking if the Status is already a generated Cert:
	if certRequest.Status.Phase == certificatev1alpha2.CertificatePhaseSigned || certRequest.Status.Phase == certificatev1alpha2.CertificatePhaseRejected {
		zap.L().Debug("Updated Cert request has already been processed", zap.String("name", certRequest.Name))
		return
	}

	if certRequest.Status.Phase == certificatev1alpha2.CertificatePhaseSubmitted {
		zap.L().Info("Processing Cert request")
		c.processCert(certRequest)
		return
	}

	// otherwise nothing has to be done
	zap.L().Info("Cert Request updated still has to be generated", zap.String("name", certRequest.Name))

}

func (c *CertificateController) onDelete(obj interface{}) {
	zap.L().Debug("Deleting Cert event")
}

// processCert is called from `onUpdate` or `onAdd` to sign/reject a CSR
func (c *CertificateController) processCert(certRequest *certificatev1alpha2.Certificate) {
	csrs, err := tglib.LoadCSRs(certRequest.Spec.Request)
	if err != nil {
		zap.L().Error("Error loading CSR", zap.Error(err), zap.String("name", certRequest.Name))
		return
	}
	if len(csrs) > 1 {
		zap.L().Error("Error loading CSR: 0 or more than 1 CSRs attached", zap.Error(err), zap.String("name", certRequest.Name), zap.Int("CSRAmount", len(csrs)))
		return
	}

	csr := csrs[0]

	zap.L().Info("Validating cert request", zap.String("name", certRequest.Name))

	err = c.issuer.Validate(csr)
	if err != nil {
		zap.L().Error("CSR has not been validated", zap.Error(err), zap.String("name", certRequest.Name))
		return
	}
	zap.L().Info("Cert request has been accepted", zap.String("name", certRequest.Name))

	cert, err := c.issuer.Sign(csr)
	if err != nil {
		zap.L().Error("Error signing CSR", zap.Error(err), zap.String("name", certRequest.Name))
		return
	}
	zap.L().Info("Cert successfully generated", zap.String("name", certRequest.Name))

	x509Cert, err := tglib.ReadCertificatePEMFromData(cert)
	if err != nil {
		zap.L().Error("Error loading x509 Cert", zap.Error(err), zap.String("name", certRequest.Name))
		return
	}

	token, err := c.issuer.IssueToken(x509Cert)
	if err != nil {
		zap.L().Error("Error Issuing compact PKI token", zap.Error(err), zap.String("name", certRequest.Name))
		return
	}

	zap.L().Debug("Cert and token successfully generated", zap.ByteString("cert", cert))

	certRequest.Status.Certificate = cert
	certRequest.Status.Ca = c.issuer.GetCACert()
	certRequest.Status.Token = token
	certRequest.Status.Phase = certificatev1alpha2.CertificatePhaseSigned

	_, err = c.certificateClient.Certificates().Update(certRequest)
	if err != nil {
		zap.L().Error("Error Updating the Certificate ressource", zap.Error(err), zap.String("name", certRequest.Name))
		return
	}
}
