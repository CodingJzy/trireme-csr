package certificates

import (
	"crypto"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/aporeto-inc/tg/tglib"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	certificatev1alpha1 "github.com/aporeto-inc/trireme-csr/apis/v1alpha1"
	certificateclient "github.com/aporeto-inc/trireme-csr/client"
)

// CertManager manages the client side for the client.
type CertManager struct {
	certName   string
	keyPass    string
	privateKey crypto.PrivateKey
	// CSR is encoded in PEM format.
	csr  []byte
	cert *x509.Certificate

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
	privateKey, err := tglib.ECPrivateKeyGenerator()
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
func (m *CertManager) GetKey() crypto.PrivateKey {
	return m.privateKey
}

// GetCert return the privateKey
func (m *CertManager) GetCert() (*x509.Certificate, error) {
	if m.cert != nil {
		return m.cert, nil
	}
	return nil, fmt.Errorf("Cert is not received yet")
}

// SendAndWaitforCert is a blocking func that issue the CertificateRequest and
// returns once the Certificate is available.
func (m *CertManager) SendAndWaitforCert(timeout time.Duration) error {
	if _, err := m.certClient.Certificates("default").Get(m.certName, metav1.GetOptions{}); err == nil {
		return fmt.Errorf("Couldn't query for existing CSR object %s", err)
	}

	kubeCert := &certificatev1alpha1.Certificate{
		Spec: certificatev1alpha1.CertificateSpec{
			Request: m.csr,
		},
	}
	kubeCert.Name = m.certName

	_, err := m.certClient.Certificates("default").Create(kubeCert)
	if err != nil {
		return fmt.Errorf("Couldn't create CSR Kube object %s", err)
	}

	watcher, err := m.certClient.Certificates("default").Watch(metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("Couldn't open a watcher for certificates %s", err)
	}
	resultChan := watcher.ResultChan()
	timeoutChan := time.After(timeout)
	for {
		select {
		case <-timeoutChan:
			return fmt.Errorf("Timed out for certificate generation")

		case event := <-resultChan:
			fmt.Printf("Received an event %+v", event)
			if event.Object != nil {
				certKube := event.Object.(*certificatev1alpha1.Certificate)
				if certKube.GetName() == m.certName {
					if certKube.Status.Certificate != nil {
						return nil
					}
				}
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
