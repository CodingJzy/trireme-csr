package certificates

import (
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
)

type certManager struct {
	keyPass    string
	privateKey []byte
	csr        []byte
}

func (m *certManager) GeneratePrivateKey() {

}

func (m *certManager) GetCertificate() {

}

func (m *certManager) GetKey() {

}

// SendAndWaitforCert is a blocking func that issue the CertificateRequest and
// returns once the Certificate is available.
func (m *certManager) SendAndWaitforCert() error {

}

// IsIssued returns true if the certificate is issued by the controller
func (m *certManager) IsIssued() (bool, error) {

}

func (m *certManager) IssueCertificate() (bool, error) {

	asn1Data, err := x509.CreateCertificate(rand.Reader, x509Cert, signerCert, pub, signerKey)
	if err != nil {
		return nil, nil, err
	}

	privPEM, err := KeyToPEM(priv)
	if err != nil {
		return nil, nil, err
	}

	certPEM := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: asn1Data,
	}

}
