package persistor

import "github.com/aporeto-inc/trireme-csr/ca"

// Interface describes the CA persistor interface
type Interface interface {
	// Store the given CA
	Store(ca ca.CertificateAuthority) error
	// Load the stored CA
	Load() (*ca.CertificateAuthority, error)
	// Delete the stored CA
	Delete() error
	// Overwrite the stored CA
	Overwrite(ca ca.CertificateAuthority) error
	// Exists returns true if there is a CA stored
	Exists() bool
}
