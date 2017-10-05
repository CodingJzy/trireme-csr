package certificates

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"

	"go.uber.org/zap"

	"github.com/aporeto-inc/tg/tglib"
)

// Issuer takes CSRs and issues valid certificates based on a valid CA
type Issuer struct {
	signingCert    *x509.Certificate
	signingKey     crypto.PrivateKey
	signingKeyPass string
}

// NewIssuer creates an issuer based on crypto CA objects
func NewIssuer(signingCert *x509.Certificate, signingKey crypto.PrivateKey, signingKeyPass string) (*Issuer, error) {
	return &Issuer{
		signingCert:    signingCert,
		signingKey:     signingKey,
		signingKeyPass: signingKeyPass,
	}, nil
}

// NewIssuerFromPath creates an issuer based on the path of PEM encoded crypto primitives
func NewIssuerFromPath(signingCertPath, signingCertKeyPath, signingKeyPass string) (*Issuer, error) {
	signingCert, signingKey, err := tglib.ReadCertificatePEM(signingCertPath, signingCertKeyPath, signingKeyPass)
	if err != nil {
		return nil, err
	}

	return NewIssuer(signingCert, signingKey, signingKeyPass)
}

// Validate verifys that the CSR is allowed to be issued. Return an error if not allowed.
func (s *Issuer) Validate(csr *x509.CertificateRequest) error {
	return nil
}

// Sign generate a signed and valid certificate for the CSR given as parameter
func (s *Issuer) Sign(csr *x509.CertificateRequest) ([]byte, error) {

	// Generate random serial number.
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		zap.L().Error("Failed to generate serial number for the certificate", zap.Error(err))
		return nil, fmt.Errorf("Failed to generate serial number for the certificate")
	}

	// Create certfificate template.
	template := x509.Certificate{
		SerialNumber:          serialNumber,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		BasicConstraintsValid: true,
		PublicKey:             csr.PublicKey,
		PublicKeyAlgorithm:    csr.PublicKeyAlgorithm,
		Subject:               csr.Subject,
		EmailAddresses:        csr.EmailAddresses,
		DNSNames:              csr.DNSNames,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, s.signingCert, csr.PublicKey, s.signingKey)
	if err != nil {
		zap.L().Error("Failed to create certificate", zap.Error(err))
		return nil, fmt.Errorf("Failed to create certificate")
	}

	clientCertificate := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})

	certificatePem := bytes.TrimSpace(clientCertificate)

	return certificatePem, nil
}
