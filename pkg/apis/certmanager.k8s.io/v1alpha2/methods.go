package v1alpha2

import (
	"crypto/x509"
	"fmt"

	"go.aporeto.io/tg/tglib"
)

// GetCertificateRequest returns a `*x509.CertificateRequest` object from the spec, or
// an error if this fails.
func (c *CertificateSpec) GetCertificateRequest() (*x509.CertificateRequest, error) {
	if c.Request == nil {
		return nil, fmt.Errorf("no certificate request in spec")
	}
	csrs, err := tglib.LoadCSRs(c.Request)
	if err != nil {
		return nil, err
	}
	if len(csrs) != 1 {
		return nil, fmt.Errorf("spec must contain exactly one CSR")
	}
	return csrs[0], nil
}

// GetCertificate returns a `*x509.Certificate` object from the status holding the
// issued certificate from the CA, or an error if this fails
func (c *CertificateStatus) GetCertificate() (*x509.Certificate, error) {
	if c.Certificate == nil {
		return nil, fmt.Errorf("no certificate has been issued yet")
	}
	x509certs, err := tglib.ParseCertificate(c.Ca)
	if err != nil {
		return nil, err
	}
	return x509certs, nil
}

// GetCACertificate returns a `*x509.Certificate` object from the status holding the issuing CA
// certificate, or an error if this fails
func (c *CertificateStatus) GetCACertificate() (*x509.Certificate, error) {
	if c.Ca == nil {
		return nil, fmt.Errorf("no CA certificate found")
	}

	return tglib.ParseCertificate(c.Ca)
}

// GetCertificateRequest returns a `*x509.CertificateRequest` object from the spec, or
// an error if this fails.
func (c *Certificate) GetCertificateRequest() (*x509.CertificateRequest, error) {
	return c.Spec.GetCertificateRequest()
}

// GetCertificate returns a `*x509.Certificate` object from the status holding the
// issued certificate from the CA, or an error if this fails
func (c *Certificate) GetCertificate() (*x509.Certificate, error) {
	return c.Status.GetCertificate()
}

// GetCACertificate returns a `*x509.Certificate` object from the status holding the issuing CA
// certificate, or an error if this fails
func (c *Certificate) GetCACertificate() (*x509.Certificate, error) {
	return c.Status.GetCACertificate()
}
