package certificates

import (
	"bytes"
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/aporeto-inc/tg/tglib"
)

// Issuer is able to validate and sign certificates baed on a CSR.
type Issuer interface {
	Validate(csr *x509.CertificateRequest) error
	Sign(csr *x509.CertificateRequest) ([]byte, error)
	GetCACert() []byte
}

// TriremeIssuer takes CSRs and issues valid certificates based on a valid CA
type TriremeIssuer struct {
	signingCert    *x509.Certificate
	signingCertPEM []byte
	signingKey     crypto.PrivateKey
}

// NewTriremeIssuer creates an issuer based on crypto CA objects
// TODO: Remove the double reference to the SigningCert.
func NewTriremeIssuer(signingCertPEM []byte, signingCert *x509.Certificate, signingKey crypto.PrivateKey, signingKeyPass string) (*TriremeIssuer, error) {

	return &TriremeIssuer{
		signingCert:    signingCert,
		signingCertPEM: signingCertPEM,
		signingKey:     signingKey,
	}, nil
}

// NewTriremeIssuerFromPath creates an issuer based on the path of PEM encoded crypto primitives
func NewTriremeIssuerFromPath(signingCertPath, signingCertKeyPath, signingKeyPass string) (*TriremeIssuer, error) {
	signingCert, signingKey, err := tglib.ReadCertificatePEM(signingCertPath, signingCertKeyPath, signingKeyPass)
	if err != nil {
		return nil, err
	}

	caCertPEM, err := LoadCertPEM(signingCertPath)
	if err != nil {
		return nil, fmt.Errorf("Error loading CaCertFile %s", err)
	}

	return NewTriremeIssuer(caCertPEM, signingCert, signingKey, signingKeyPass)
}

// Validate verifys that the CSR is allowed to be issued. Return an error if not allowed.
func (i *TriremeIssuer) Validate(csr *x509.CertificateRequest) error {
	return nil
}

// Sign generate a signed and valid certificate for the CSR given as parameter
func (i *TriremeIssuer) Sign(csr *x509.CertificateRequest) ([]byte, error) {

	pemCert, _, err := tglib.SignCSR(csr,
		i.signingCert,
		i.signingKey,
		time.Now(),
		time.Now().AddDate(1, 0, 0),
		x509.KeyUsageDigitalSignature|x509.KeyUsageKeyEncipherment,
		[]x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		x509.ECDSAWithSHA384,
		x509.ECDSA,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to generate Cert: %s", err)
	}

	clientCertificate := pem.EncodeToMemory(pemCert)

	certificatePem := bytes.TrimSpace(clientCertificate)

	return certificatePem, nil
}

// GetCACert returns the CA Certificate that is used for this issuer.
func (i *TriremeIssuer) GetCACert() []byte {
	return i.signingCertPEM
}

// LoadCertPEM returns the byte array of the PEM encoded Cert.
func LoadCertPEM(signingCertPath string) ([]byte, error) {
	certPemBytes, err := ioutil.ReadFile(signingCertPath)
	if err != nil {
		return nil, fmt.Errorf("Unable to read certificate %s", err)
	}

	return certPemBytes, nil
}
