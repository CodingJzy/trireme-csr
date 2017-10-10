package certificates

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/aporeto-inc/tg/tglib"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	certificatev1alpha1 "github.com/aporeto-inc/trireme-csr/apis/v1alpha1"
	certificateclient "github.com/aporeto-inc/trireme-csr/client"
)

// CertManager manages the client side for the client.
// It encapsulates the PrivateKey that should always remain private to this pod.
type CertManager struct {
	certName   string
	keyPass    string
	privateKey *ecdsa.PrivateKey
	// CSR is encoded in PEM format.
	csr []byte

	certPEM []byte
	cert    *x509.Certificate

	caCertPEM []byte
	caCert    *x509.Certificate

	certClient *certificateclient.CertificateClient
}

// NewCertManager creates a NewCertManager with default.
func NewCertManager(name string, certClient *certificateclient.CertificateClient) (*CertManager, error) {
	return &CertManager{
		certName:   name,
		certClient: certClient,
	}, nil
}

// GeneratePrivateKey generate the private key that will be used for this Certificate.
func (m *CertManager) GeneratePrivateKey() error {
	privateKey, err := ECPrivateKeyGenerator()
	m.privateKey = privateKey
	if err != nil {
		return err
	}
	return nil
}

// GenerateCSR generates the CSR associated with the key
func (m *CertManager) GenerateCSR() error {
	oidEmailAddress := asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 1}
	emailAddress := "aporeto@aporeto.com"
	subj := pkix.Name{
		CommonName:         "enforcerd",
		Country:            []string{"US"},
		Province:           []string{"Some-State"},
		Locality:           []string{"MyCity"},
		Organization:       []string{"Company Ltd"},
		OrganizationalUnit: []string{"IT"},
	}
	rawSubj := subj.ToRDNSequence()
	rawSubj = append(rawSubj, []pkix.AttributeTypeAndValue{
		{Type: oidEmailAddress, Value: emailAddress},
	})

	asn1Subj, _ := asn1.Marshal(rawSubj)
	certRequest := x509.CertificateRequest{
		RawSubject:         asn1Subj,
		EmailAddresses:     []string{emailAddress},
		SignatureAlgorithm: x509.ECDSAWithSHA256,
	}

	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &certRequest, m.privateKey)
	if err != nil {
		return fmt.Errorf("Couldn't create CSR %s", err)
	}

	m.csr = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrBytes})

	return nil
}

// GetKey return the privateKey
func (m *CertManager) GetKey() *ecdsa.PrivateKey {
	return m.privateKey
}

// GetKeyPEM return the privateKey in PEM format
func (m *CertManager) GetKeyPEM() ([]byte, error) {
	keybyte, err := x509.MarshalECPrivateKey(m.privateKey)
	if err != nil {
		return nil, fmt.Errorf("Error marshaling ECDSA Private Key %s", err)
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keybyte})
	return keyPEM, nil
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

// GetCaCert return the privateKey
func (m *CertManager) GetCaCert() (*x509.Certificate, error) {
	if m.caCert == nil {
		return nil, fmt.Errorf("Cert is not received yet")
	}

	return m.caCert, nil
}

// GetCaCertPEM return the privateKey in PEM format
func (m *CertManager) GetCaCertPEM() ([]byte, error) {
	if m.caCertPEM == nil {
		return nil, fmt.Errorf("Cert is not received yet")
	}

	return m.caCertPEM, nil
}

// SendAndWaitforCert is a blocking func that issue the CertificateRequest and
// returns once the Certificate is available.
func (m *CertManager) SendAndWaitforCert(timeout time.Duration) error {

	// First check if the certificate was already issued.
	certs, err := m.certClient.Certificates("default").List(metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("Couldn't query for existing CSR object list %s", err)
	}
	for _, cert := range certs.Items {
		if cert.Name == m.certName {
			if cert.Status.Certificate != nil {
				return nil
			}
		}
	}

	// Generate the new certificate kube object
	kubeCert := &certificatev1alpha1.Certificate{
		Spec: certificatev1alpha1.CertificateSpec{
			Request: m.csr,
		},
	}
	kubeCert.Name = m.certName

	zap.L().Info("Creating new certificate object on Kube API", zap.String("certName", m.certName))
	_, err = m.certClient.Certificates("default").Create(kubeCert)
	if err != nil {
		return fmt.Errorf("Couldn't create CSR Kube object %s", err)
	}

	timeoutChan := time.After(timeout)
	tickerChan := time.NewTicker(time.Second).C

	for {
		select {
		case <-timeoutChan:
			return fmt.Errorf("Timed out for certificate generation")

		case <-tickerChan:
			zap.L().Info("Verifying if Certificate was issued by controller...", zap.String("certName", m.certName))
			cert, err := m.certClient.Certificates("default").Get(m.certName, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("Existing CSR object deleted %s", err)
			}

			if cert.Status.Certificate != nil {
				fmt.Printf("Cert is available: %+v", cert.Status.Certificate)
				m.certPEM = cert.Status.Certificate
				m.cert, err = tglib.ReadCertificatePEMFromData(cert.Status.Certificate)
				if err != nil {
					return fmt.Errorf("Couldn't parse certificate %s", err)
				}

				m.caCertPEM = cert.Status.Ca
				m.caCert, err = tglib.ReadCertificatePEMFromData(cert.Status.Certificate)
				if err != nil {
					return fmt.Errorf("Couldn't parse CA certificate %s", err)
				}

				return nil
			}
		}
	}
}

// IsIssued returns true if the certificate is issued by the controller
func (m *CertManager) IsIssued() bool {
	if m.cert != nil {
		return true
	}
	return false
}

// ECPrivateKeyGenerator generates a ECDSA private key.
func ECPrivateKeyGenerator() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}
