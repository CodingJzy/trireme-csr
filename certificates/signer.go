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

type Signer struct {
	signingCert    *x509.Certificate
	signingKey     crypto.PrivateKey
	signingKeyPass string
}

func NewSigner(signingCert *x509.Certificate, signingKey crypto.PrivateKey, signingKeyPass string) (*Signer, error) {
	return &Signer{
		signingCert:    signingCert,
		signingKey:     signingKey,
		signingKeyPass: signingKeyPass,
	}, nil
}

func NewSignerFromPath(signingCertPath, signingCertKeyPath, signingKeyPass string) (*Signer, error) {
	signingCert, signingKey, err := tglib.ReadCertificatePEM(signingCertPath, signingCertKeyPath, signingKeyPass)
	if err != nil {
		return nil, err
	}

	return &Signer{
		signingCert:    signingCert,
		signingKey:     signingKey,
		signingKeyPass: signingKeyPass,
	}, nil
}

func (s *Signer) Validate(csr *x509.CertificateRequest) error {
	return nil
}

// Sign generate a signed and valid certificate for the CSR given as parameter
func (s *Signer) Sign(csr *x509.CertificateRequest) ([]byte, error) {

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

// LoadCSR loads the given bytes as a Certificate Signing Request.
func LoadCSR(csrData []byte) (*x509.CertificateRequest, error) {

	block, _ := pem.Decode(csrData)
	if block == nil {
		return nil, fmt.Errorf("Given CSR is not a valid PEM")
	}

	csr, err := x509.ParseCertificateRequest(block.Bytes)
	if err != nil {
		return nil, err
	}

	if err := csr.CheckSignature(); err != nil {
		return nil, err
	}

	return csr, nil
}
