package certificates

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/aporeto-inc/tg/tglib"
	"github.com/aporeto-inc/trireme-lib/enforcer/utils/pkiverifier"
)

// Issuer is able to validate and sign certificates baed on a CSR.
type Issuer interface {
	Validate(csr *x509.CertificateRequest) error
	Sign(csr *x509.CertificateRequest) ([]byte, error)
	IssueToken(cert *x509.Certificate) ([]byte, error)
	GetCACert() []byte
}

// TriremeIssuer takes CSRs and issues valid certificates based on a valid CA
type TriremeIssuer struct {
	signingCert    *x509.Certificate
	signingCertPEM []byte
	signingKey     crypto.PrivateKey

	tokenIssuer pkiverifier.PKITokenIssuer
}

// NewTriremeIssuer creates an issuer based on crypto CA objects
// TODO: Remove the double reference to the SigningCert.
func NewTriremeIssuer(signingCertPEM []byte, signingCert *x509.Certificate, signingKey crypto.PrivateKey, signingKeyPass string) (*TriremeIssuer, error) {
	// TODO: Bettre validation of parameters here.
	pkiIssuer := pkiverifier.NewPKIIssuer(signingKey.(*ecdsa.PrivateKey))

	return &TriremeIssuer{
		signingCert:    signingCert,
		signingCertPEM: signingCertPEM,
		signingKey:     signingKey,

		tokenIssuer: pkiIssuer,
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

	var keyUsage x509.KeyUsage
	var extKeyUsage []x509.ExtKeyUsage

	// TODO: Revisit the existing KeyUsage.
	keyUsage = x509.KeyUsageDigitalSignature
	keyUsage |= x509.KeyUsageKeyEncipherment

	extKeyUsage = append(extKeyUsage, x509.ExtKeyUsageClientAuth)
	extKeyUsage = append(extKeyUsage, x509.ExtKeyUsageServerAuth)

	pemCert, _, err := tglib.SignCSR(csr,
		i.signingCert,
		i.signingKey,
		time.Now(),
		time.Now().AddDate(1, 0, 0),
		keyUsage,
		extKeyUsage,
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

// IssueToken generates a valid token for the cert given as parameter
func (i *TriremeIssuer) IssueToken(cert *x509.Certificate) ([]byte, error) {
	return i.tokenIssuer.CreateTokenFromCertificate(cert)
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
