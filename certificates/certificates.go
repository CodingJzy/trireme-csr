package certificates

import (
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"

	"go.uber.org/zap"

	"go.aporeto.io/tg/tglib"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	certificatev1alpha2 "github.com/aporeto-inc/trireme-csr/pkg/apis/certmanager.k8s.io/v1alpha2"
	certificateclient "github.com/aporeto-inc/trireme-csr/pkg/client/clientset/versioned"
)

// CertManager manages the client side for the client.
// It encapsulates the PrivateKey that should always remain private to this pod.
type CertManager struct {
	certName   string
	keyPass    string
	privateKey crypto.PrivateKey
	// CSR is encoded in PEM format.
	csr []byte

	certPEM []byte
	cert    *x509.Certificate

	caCertPEM []byte
	caCert    *x509.Certificate

	smartToken []byte

	certClient certificateclient.Interface
}

// NewCertManager creates a NewCertManager with default.
func NewCertManager(name string, certClient certificateclient.Interface) (*CertManager, error) {
	return &CertManager{
		certName:   name,
		certClient: certClient,
	}, nil
}

// GeneratePrivateKey generate the private key that will be used for this Certificate.
func (m *CertManager) GeneratePrivateKey() error {
	privateKey, err := tglib.ECPrivateKeyGenerator()
	m.privateKey = privateKey
	if err != nil {
		return err
	}
	return nil
}

// GenerateCSR generates the CSR associated with the key
func (m *CertManager) GenerateCSR() error {
	emailAddress := "aporeto@aporeto.com"

	certRequest, err := tglib.GenerateSimpleCSR(
		[]string{"trireme"},
		[]string{"unit"},
		"commonName",
		[]string{emailAddress},
		m.privateKey,
	)
	if err != nil {
		return err
	}

	m.csr = certRequest
	return nil
}

// GetKey return the privateKey
func (m *CertManager) GetKey() crypto.PrivateKey {
	return m.privateKey
}

// GetKeyPEM return the privateKey in PEM format
func (m *CertManager) GetKeyPEM() ([]byte, error) {
	keyPEM, err := tglib.KeyToPEM(m.privateKey)
	if err != nil {
		return nil, fmt.Errorf("Error marshaling Private Key %s", err)
	}

	keyByte := pem.EncodeToMemory(keyPEM)
	return keyByte, nil
}

// GetCert return the privateKey
func (m *CertManager) GetCert() (*x509.Certificate, error) {
	if m.cert == nil {
		return nil, fmt.Errorf("Cert is not received yet")
	}
	return m.cert, nil
}

// GetCertPEM return the privateKey in PEM format
func (m *CertManager) GetCertPEM() ([]byte, error) {
	if m.cert == nil {
		return nil, fmt.Errorf("Cert is not received yet")
	}

	return m.certPEM, nil
}

// GetCaCert returns the privateKey
func (m *CertManager) GetCaCert() (*x509.Certificate, error) {
	if m.caCert == nil {
		return nil, fmt.Errorf("Cert is not received yet")
	}

	return m.caCert, nil
}

// GetCaCertPEM returns the privateKey in PEM format
func (m *CertManager) GetCaCertPEM() ([]byte, error) {
	if m.caCertPEM == nil {
		return nil, fmt.Errorf("Cert is not received yet")
	}

	return m.caCertPEM, nil
}

// GetSmartToken returns the GetSmartToken
func (m *CertManager) GetSmartToken() ([]byte, error) {
	if m.smartToken == nil {
		return nil, fmt.Errorf("SmartToken is not received yet")
	}

	return m.smartToken, nil
}

// SendAndWaitforCert is a blocking func that issue the CertificateRequest and
// returns once the Certificate is available.
func (m *CertManager) SendAndWaitforCert(timeout time.Duration) error {

	// First check if the certificate was already issued.
	certs, err := m.certClient.CertmanagerV1alpha2().Certificates().List(metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("Couldn't query for existing CSR object list: %s", err.Error())
	}
	for _, cert := range certs.Items {
		if cert.Name == m.certName {
			if err = m.certClient.CertmanagerV1alpha2().Certificates().Delete(cert.Name, &metav1.DeleteOptions{}); err != nil {
				return fmt.Errorf("Error deleting existing cert for node: %s", err.Error())
			}
		}
	}

	// Generate the new certificate kube object
	kubeCert := &certificatev1alpha2.Certificate{
		Spec: certificatev1alpha2.CertificateSpec{
			Request: m.csr,
		},
	}
	kubeCert.Name = m.certName

	zap.L().Info("Creating new certificate object on Kube API", zap.String("certName", m.certName))
	_, err = m.certClient.CertmanagerV1alpha2().Certificates().Create(kubeCert)
	if err != nil {
		return fmt.Errorf("couldn't create CSR Kube object: %s", err.Error())
	}

	timeoutChan := time.After(timeout)
	tickerChan := time.NewTicker(time.Second).C

	for {
		select {
		case <-timeoutChan:
			return fmt.Errorf("Timed out for certificate generation")

		case <-tickerChan:
			zap.L().Info("Verifying if Certificate was issued by controller...", zap.String("certName", m.certName))
			cert, err := m.certClient.CertmanagerV1alpha2().Certificates().Get(m.certName, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("existing CSR object deleted %s", err.Error())
			}

			switch cert.Status.Phase {
			case certificatev1alpha2.CertificateRejected:
				return fmt.Errorf("Certificate issuing has been rejected by the controller: %s: %s", cert.Status.Reason, cert.Status.Message)

			case certificatev1alpha2.CertificateUnknown:
				return fmt.Errorf("The controller did not know how to handle our request and moved it to the '%s' phase (%s: %s)", certificatev1alpha2.CertificateUnknown, cert.Status.Reason, cert.Status.Message)

			case certificatev1alpha2.CertificateSubmitted:
				zap.L().Sugar().Debugf("Controller has accepted our request and moved it to the '%s' phase", certificatev1alpha2.CertificateSubmitted)
				break

			case certificatev1alpha2.CertificateSigned:
				if cert.Status.Certificate != nil {
					fmt.Printf("Cert is available: %+v", cert.Status.Certificate)
					m.certPEM = cert.Status.Certificate
					m.cert, err = tglib.ReadCertificatePEMFromData(cert.Status.Certificate)
					if err != nil {
						return fmt.Errorf("couldn't parse certificate: %s", err.Error())
					}

					m.caCertPEM = cert.Status.Ca
					m.caCert, err = tglib.ReadCertificatePEMFromData(cert.Status.Certificate)
					if err != nil {
						return fmt.Errorf("couldn't parse CA certificate: %s", err.Error())
					}

					m.smartToken = cert.Status.Token
					return nil
				}

			default:
				zap.L().Debug("Unhandled certificate status phase", zap.String("phase", string(cert.Status.Phase)))
			}
		}
	}
}
