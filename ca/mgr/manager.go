package manager

import (
	"fmt"
	"sync"

	"github.com/aporeto-inc/trireme-csr/ca"
	"github.com/aporeto-inc/trireme-csr/ca/persistor"
)

// Manager struct
type Manager struct {
	lock      sync.Mutex
	ca        *ca.CertificateAuthority
	persistor persistor.Interface
}

// NewManager creates a new CA Manager
func NewManager(persistor persistor.Interface) (*Manager, error) {
	if persistor == nil {
		return nil, fmt.Errorf("must be initialized with CA persistor")
	}
	return &Manager{
		persistor: persistor,
	}, nil
}

// IsCALoaded is true if a CA is loaded into the manager
func (m *Manager) IsCALoaded() bool {
	m.lock.Lock()
	defer m.lock.Unlock()

	return m.isCALoaded()
}

// UnloadCA unloads a loaded CA
func (m *Manager) UnloadCA() {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.ca = nil
}

// LoadCAFromFiles tries to load a CA from the given certPath, keyPath and key password.
// It returns with an error if this operation fails or a CA is already loaded.
func (m *Manager) LoadCAFromFiles(certPath, keyPath, password string) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.isCALoaded() {
		return fmt.Errorf("CA is already loaded")
	}

	ca, err := ca.LoadCertificateAuthorityFromFiles(certPath, keyPath, password)
	if err != nil {
		return err
	}
	m.ca = ca
	return nil
}

// GenerateCA tries to generate a CA. It returns an error if this operation fails
// or a CA is already loaded
func (m *Manager) GenerateCA() error {
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.isCALoaded() {
		return fmt.Errorf("CA is already loaded")
	}

	ca, err := ca.NewCertificateAuthority()
	if err != nil {
		return err
	}
	m.ca = ca
	return nil
}

// isCALoaded is the internal version of IsCALoaded but just without the lock
func (m *Manager) isCALoaded() bool {
	return m.ca != nil
}
