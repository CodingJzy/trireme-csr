// Package persistor defines an interface for persisting a CA to a storage of some kind. This
// interface needs to be implemented by every persistor to be usable in the CA manager. It is
// on purpose a very high-level interface. In the future the `CertificateAuthority` will most
// likely turn into an interface itself to provide a better abstraction layer for CA persistence
// interfaces that can not directly pass a key/cert along due to their nature (like an HSM).
//
// There is currently only one implementation for Trireme-CSR: the `SecretsPersistor` which
// stores/loads a CA in a Kubernetes Secret.
package persistor

import "github.com/aporeto-inc/trireme-csr/ca"

// Interface describes the CA persistor interface. It provides all necessary functions to
// persist and load a CA from the backend implementation.
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
