package controller

import (
	"bytes"
	"crypto/x509"
	"fmt"

	"go.uber.org/zap"

	"github.com/CodingJzy/trireme-csr/certificates"
	"go.aporeto.io/tg/tglib"

	"k8s.io/client-go/tools/cache"

	certificatev1alpha2 "github.com/CodingJzy/trireme-csr/pkg/apis/certmanager.k8s.io/v1alpha2"
	certificateclient "github.com/CodingJzy/trireme-csr/pkg/client/clientset/versioned"
	certificateinformers "github.com/CodingJzy/trireme-csr/pkg/client/informers/externalversions"
	certificateinformerv1alpha2 "github.com/CodingJzy/trireme-csr/pkg/client/informers/externalversions/certmanager.k8s.io/v1alpha2"
)

// CertificateController contains all the logic to implement the issuance of certificates.
type CertificateController struct {
	certificateClient   certificateclient.Interface
	certificateInformer certificateinformerv1alpha2.CertificateInformer
	issuer              certificates.Issuer
}

// NewCertificateController generates the new CertificateController
func NewCertificateController(certificateClient certificateclient.Interface, certificateInformerFactory certificateinformers.SharedInformerFactory, issuer certificates.Issuer) *CertificateController {

	certificateInformer := certificateInformerFactory.Certmanager().V1alpha2().Certificates()

	c := &CertificateController{
		certificateClient:   certificateClient,
		certificateInformer: certificateInformer,
		issuer:              issuer,
	}

	certificateInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    c.onAdd,
			UpdateFunc: c.onUpdate,
			DeleteFunc: c.onDelete,
		},
	)

	return c
}

// Run starts the certificateWatcher.
func (c *CertificateController) Run(stopCh <-chan struct{}) error {
	zap.L().Info("start watching Certificates objects")

	// wait for caches to sync
	ok := cache.WaitForCacheSync(stopCh, c.certificateInformer.Informer().HasSynced)
	if !ok {
		return fmt.Errorf("error while waiting for caches to sync")
	}

	// now wait until the stopCh closes
	<-stopCh
	return nil
}

func (c *CertificateController) onAdd(obj interface{}) {
	certRequest, ok := obj.(*certificatev1alpha2.Certificate)
	if !ok {
		zap.L().Sugar().Errorf("Received wrong object type in adding Cert event: '%T", obj)
		return
	}

	switch certRequest.Status.Phase {
	case certificatev1alpha2.CertificateSigned:
		zap.L().Debug("Added Cert request has already been processed and a certificate was issued", zap.String("name", certRequest.Name), zap.String("resource_version", certRequest.ResourceVersion))
		// if a cert is added as signed, we should validate it before we allow this status like this from scratch
		// 1. check if the CSR was actually valid
		csr, err := certRequest.GetCertificateRequest()
		if err != nil {
			c.updateCertRejected(
				certRequest,
				certificatev1alpha2.StatusReasonProcessedRejectedInvalidCSR,
				fmt.Errorf("changing phase to '%s': failed to get CSR: %s", certificatev1alpha2.CertificateRejected, err.Error()),
			)
			return
		}
		err = c.issuer.ValidateRequest(csr)
		if err != nil {
			c.updateCertRejected(
				certRequest,
				certificatev1alpha2.StatusReasonProcessedRejectedInvalidCSR,
				fmt.Errorf("changing phase to '%s': failed to validate CSR: %s", certificatev1alpha2.CertificateRejected, err.Error()),
			)
			return
		}
		// 2. check if the issued cert is valid
		cert, err := certRequest.GetCertificate()
		if err != nil {
			c.updateCertRejected(
				certRequest,
				certificatev1alpha2.StatusReasonProcessedRejectedInvalidCerts,
				fmt.Errorf("changing phase to '%s': failed to get Certificate: %s", certificatev1alpha2.CertificateRejected, err.Error()),
			)
			return
		}
		// as this has been added as a new object, we retrieve the CA from the object here for validation,
		// as chances are that the CA cert is not the cert of the running CA
		ca, err := certRequest.GetCACertificate()
		if err != nil {
			c.updateCertRejected(
				certRequest,
				certificatev1alpha2.StatusReasonProcessedRejectedInvalidCerts,
				fmt.Errorf("changing phase to '%s': failed to get CA Certificate: %s", certificatev1alpha2.CertificateRejected, err.Error()),
			)
			return
		}
		err = c.issuer.ValidateCert(cert, ca)
		if err != nil {
			c.updateCertRejected(
				certRequest,
				certificatev1alpha2.StatusReasonProcessedRejectedInvalidCerts,
				fmt.Errorf("changing phase to '%s': failed to validate signed certificate: %s", certificatev1alpha2.CertificateRejected, err.Error()),
			)
		}
		// it is a valid object, nothing more to be done

	case certificatev1alpha2.CertificateRejected:
		zap.L().Debug("Added Cert request has already been processed and was rejected", zap.String("name", certRequest.Name), zap.String("resource_version", certRequest.ResourceVersion))

	case certificatev1alpha2.CertificateSubmitted:
		zap.L().Info("Added Cert request: processing Cert request", zap.String("name", certRequest.Name), zap.String("resource_version", certRequest.ResourceVersion))
		// we definitely want to process this request when the object has been created with the Submitted Phase
		c.process(certRequest)

	case certificatev1alpha2.CertificateUnknown:
		// nothing has to be done when we are in the 'Unknown' phase at Adding time
		zap.L().Debug("Added Cert request: nothing has to be done for this Cert request", zap.String("name", certRequest.Name), zap.String("resource_version", certRequest.ResourceVersion))

	default:
		// this means that a phase is missing, which should be the default when one creates an object
		// check if we have a spec, if yes, move it to the submitted phase
		if certRequest.Spec.Request != nil && len(certRequest.Spec.Request) > 0 {
			zap.L().Debug("Added Cert request: phase missing, but spec exists -> got just submitted", zap.String("name", certRequest.Name), zap.String("resource_version", certRequest.ResourceVersion))
			c.updateCertSubmitted(certRequest)
			return
		}

		// there is no spec, so we move it to the Unknown phase
		// the user will have to add a spec, so that it can get moved into the submitted phase
		zap.L().Debug("Added Cert request: phase missing, but no spec -> unrecognized phase", zap.String("name", certRequest.Name), zap.String("resource_version", certRequest.ResourceVersion))
		c.updateCertUnknown(certRequest)
	}
}

func (c *CertificateController) onUpdate(oldObj, newObj interface{}) {
	certRequest, ok := newObj.(*certificatev1alpha2.Certificate)
	if !ok {
		zap.L().Sugar().Errorf("Received wrong object type in updating Cert event for new object: '%T", newObj)
		return
	}
	oldCertRequest, ok := oldObj.(*certificatev1alpha2.Certificate)
	if !ok {
		zap.L().Sugar().Errorf("Received wrong object type in updating Cert event for old object: '%T", oldObj)
		return
	}

	// the periodic resync of the controller will send update events
	// a different resource version of the same certificate means that
	// there was a real update
	if certRequest.ResourceVersion == oldCertRequest.ResourceVersion {
		zap.L().Debug("Updating Cert request: no change in resource version -> no real update", zap.String("name", certRequest.Name), zap.String("resource_version", certRequest.ResourceVersion))
		return
	}

	switch certRequest.Status.Phase {
	case certificatev1alpha2.CertificateSigned:
		zap.L().Debug("Updated Cert request has been processed and a certificate was issued", zap.String("name", certRequest.Name), zap.String("resource_version", certRequest.ResourceVersion))
		// Signed is a final state in the update phase
		// the only thing that we will do is to validate the certs again, to ensure that this is not a rogue update
		cert, err := certRequest.GetCertificate()
		if err != nil {
			c.updateCertRejected(
				certRequest,
				certificatev1alpha2.StatusReasonProcessedRejectedInvalidCerts,
				fmt.Errorf("changing phase to '%s': failed to get Certificate: %s", certificatev1alpha2.CertificateRejected, err.Error()),
			)
			return
		}
		err = c.issuer.ValidateCert(cert, nil)
		if err != nil {
			c.updateCertRejected(
				certRequest,
				certificatev1alpha2.StatusReasonProcessedRejectedInvalidCerts,
				fmt.Errorf("changing phase to '%s': failed to validate signed certificate: %s", certificatev1alpha2.CertificateRejected, err.Error()),
			)
		}
		// it is a valid object, nothing more to be done

	case certificatev1alpha2.CertificateRejected:
		zap.L().Debug("Update Cert request has been processed and was rejected", zap.String("name", certRequest.Name), zap.String("resource_version", certRequest.ResourceVersion))
		// check if a new spec was submitted, and move to the submitted phase if yes
		if certRequest.Spec.Request != nil && len(certRequest.Spec.Request) > 0 && !bytes.Equal(oldCertRequest.Spec.Request, certRequest.Spec.Request) {
			c.updateCertSubmitted(certRequest)
		}

	case certificatev1alpha2.CertificateUnknown:
		zap.L().Debug("Updated Cert request: nothing has to be done for this Cert request", zap.String("name", certRequest.Name), zap.String("resource_version", certRequest.ResourceVersion))
		// check if a spec has been submitted, and move it to the submitted phase if yes
		if certRequest.Spec.Request != nil && len(certRequest.Spec.Request) > 0 && !bytes.Equal(oldCertRequest.Spec.Request, certRequest.Spec.Request) {
			c.updateCertSubmitted(certRequest)
		}

	case certificatev1alpha2.CertificateSubmitted:
		zap.L().Info("Updated Cert request: processing Cert request", zap.String("name", certRequest.Name), zap.String("resource_version", certRequest.ResourceVersion))
		c.process(certRequest)

	default:
		// this means that a phase is missing, so we move it to the Unknown phase
		// as we honestly don't understand how we ended up in this state :)
		zap.L().Warn("Updated Cert request: unrecognized phase", zap.String("name", certRequest.Name), zap.String("resource_version", certRequest.ResourceVersion))
		c.updateCertUnknown(certRequest)
	}
}

func (c *CertificateController) onDelete(obj interface{}) {
	certRequest, ok := obj.(*certificatev1alpha2.Certificate)
	if !ok {
		zap.L().Sugar().Errorf("Received wrong object type in adding Cert event: '%T", obj)
		return
	}
	zap.L().Debug("Deleting Cert event", zap.String("name", certRequest.Name), zap.String("resource_version", certRequest.ResourceVersion))
}

// process is called from a `Submitted` phase event from `onUpdate` or `onAdd` to process the request
func (c *CertificateController) process(certRequest *certificatev1alpha2.Certificate) {
	// Load CSR
	csr, err := certRequest.GetCertificateRequest()
	if err != nil {
		zap.L().Error("Error loading CSR", zap.Error(err), zap.String("name", certRequest.Name), zap.String("resource_version", certRequest.ResourceVersion))
		c.updateCertRejected(
			certRequest,
			certificatev1alpha2.StatusReasonProcessedRejectedInvalidCSR,
			fmt.Errorf("Error loading CSR: %s", err.Error()),
		)
		return
	}

	// Validate CSR
	zap.L().Info("Validating cert request", zap.String("name", certRequest.Name), zap.String("resource_version", certRequest.ResourceVersion))
	err = c.issuer.ValidateRequest(csr)
	if err != nil {
		zap.L().Error("CSR has not been validated", zap.Error(err), zap.String("name", certRequest.Name), zap.String("resource_version", certRequest.ResourceVersion))
		c.updateCertRejected(
			certRequest,
			certificatev1alpha2.StatusReasonProcessedRejectedInvalidCSR,
			fmt.Errorf("Failed to validate CSR: %s", err.Error()),
		)
		return
	}
	zap.L().Info("Cert request has been accepted", zap.String("name", certRequest.Name), zap.String("resource_version", certRequest.ResourceVersion))

	// Check Key type: currently it *must* be an ECDSA key
	if csr.PublicKeyAlgorithm != x509.ECDSA {
		zap.L().Error("CSR is generated from an unsupported key type", zap.String("name", certRequest.Name), zap.String("resource_version", certRequest.ResourceVersion))
		c.updateCertRejected(
			certRequest,
			certificatev1alpha2.StatusReasonProcessedRejectedInvalidCSR,
			fmt.Errorf("Unsupported Key Type (only ECDSA keys are supported)"),
		)
		return
	}

	// Sign CSR
	cert, err := c.issuer.Sign(csr)
	if err != nil {
		zap.L().Error("Error signing CSR", zap.Error(err), zap.String("name", certRequest.Name), zap.String("resource_version", certRequest.ResourceVersion))
		c.updateCertRejected(
			certRequest,
			certificatev1alpha2.StatusReasonProcessedRejected,
			fmt.Errorf("Failed to sign CSR: %s", err.Error()),
		)
		return
	}
	zap.L().Info("Cert successfully generated", zap.String("name", certRequest.Name), zap.String("resource_version", certRequest.ResourceVersion))

	// Load the certificate as x509.Certificate as well, so that we can issue the token
	x509Cert, err := tglib.ReadCertificatePEMFromData(cert)
	if err != nil {
		zap.L().Error("Error loading x509 Cert", zap.Error(err), zap.String("name", certRequest.Name), zap.String("resource_version", certRequest.ResourceVersion))
		c.updateCertRejected(
			certRequest,
			certificatev1alpha2.StatusReasonProcessedRejected,
			fmt.Errorf("Error loading x509 Cert: %s", err.Error()),
		)
		return
	}

	// issue token
	token, err := c.issuer.IssueToken(x509Cert)
	if err != nil {
		zap.L().Error("Error Issuing compact PKI token", zap.Error(err), zap.String("name", certRequest.Name), zap.String("resource_version", certRequest.ResourceVersion))
		c.updateCertRejected(
			certRequest,
			certificatev1alpha2.StatusReasonProcessedRejected,
			fmt.Errorf("Error Issuing compact PKI token: %s", err.Error()),
		)
		return
	}

	zap.L().Debug("Cert and token successfully generated", zap.String("name", certRequest.Name), zap.String("resource_version", certRequest.ResourceVersion), zap.ByteString("cert", cert))

	// last but not least, update our object with the signed cert
	c.updateCertSigned(certRequest, cert, token)
}

func (c *CertificateController) updateCertSubmitted(certRequestObj *certificatev1alpha2.Certificate) {
	certRequest := certRequestObj.DeepCopy()
	certRequest.Status.Phase = certificatev1alpha2.CertificateSubmitted
	certRequest.Status.Reason = certificatev1alpha2.StatusReasonSubmitted
	certRequest.Status.Message = "The request contains a certificate request. Submitting certificate request for processing."

	// we can not use UpdateStatus until this issue gets resolved afaik
	// https://github.com/kubernetes/kubernetes/issues/38113
	_, err := c.certificateClient.CertmanagerV1alpha2().Certificates().Update(certRequest)
	if err != nil {
		zap.L().Error("Error Updating the Certificate ressource", zap.Error(err), zap.String("name", certRequest.Name), zap.String("resource_version", certRequest.ResourceVersion))
		return
	}
}

func (c *CertificateController) updateCertUnknown(certRequestObj *certificatev1alpha2.Certificate) {
	certRequest := certRequestObj.DeepCopy()
	certRequest.Status.Phase = certificatev1alpha2.CertificateUnknown
	certRequest.Status.Reason = certificatev1alpha2.StatusReasonUnprocessed
	certRequest.Status.Message = "The request has not been processed by the controller yet. Submit a valid CSR in the spec to submit this CSR for processing."

	// we can not use UpdateStatus until this issue gets resolved afaik
	// https://github.com/kubernetes/kubernetes/issues/38113
	_, err := c.certificateClient.CertmanagerV1alpha2().Certificates().Update(certRequest)
	if err != nil {
		zap.L().Error("Error Updating the Certificate ressource", zap.Error(err), zap.String("name", certRequest.Name), zap.String("resource_version", certRequest.ResourceVersion))
		return
	}
}

func (c *CertificateController) updateCertRejected(certRequestObj *certificatev1alpha2.Certificate, reason string, rejectErr error) {
	certRequest := certRequestObj.DeepCopy()
	certRequest.Status.Phase = certificatev1alpha2.CertificateRejected
	certRequest.Status.Reason = reason
	certRequest.Status.Message = rejectErr.Error()

	// we can not use UpdateStatus until this issue gets resolved afaik
	// https://github.com/kubernetes/kubernetes/issues/38113
	_, err := c.certificateClient.CertmanagerV1alpha2().Certificates().Update(certRequest)
	if err != nil {
		zap.L().Error("Error Updating the Certificate ressource", zap.Error(err), zap.String("name", certRequest.Name), zap.String("resource_version", certRequest.ResourceVersion))
		return
	}
}

// updateCertSigned is called when a request has been successfully processed/approved/signed
func (c *CertificateController) updateCertSigned(certRequestObj *certificatev1alpha2.Certificate, cert, token []byte) {
	certRequest := certRequestObj.DeepCopy()
	certRequest.Status.Certificate = cert
	certRequest.Status.Ca = c.issuer.GetCACert()
	certRequest.Status.Token = token
	certRequest.Status.Phase = certificatev1alpha2.CertificateSigned
	certRequest.Status.Reason = certificatev1alpha2.StatusReasonProcessedApprovedSignedIssued
	certRequest.Status.Message = "CSR has been processed and approved, and the Certificate has been signed and issued"

	// we can not use UpdateStatus until this issue gets resolved afaik
	// https://github.com/kubernetes/kubernetes/issues/38113
	_, err := c.certificateClient.CertmanagerV1alpha2().Certificates().Update(certRequest)
	if err != nil {
		zap.L().Error("Error Updating the Certificate ressource", zap.Error(err), zap.String("name", certRequest.Name), zap.String("resource_version", certRequest.ResourceVersion))
		return
	}
}
