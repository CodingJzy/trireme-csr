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
// *NOTE:* This will **not** persist the CA, but only make it available to the manager!
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

// LoadCAFromPersistor tries to load the CA from ther persistance interface.
// It returns with an error if this operation fails (e.g. the stored CA is corrupt,
// or does not exist) or a CA is already loaded. *NOTE:* This will **not** persist the CA,
// but only make it available to the manager!
func (m *Manager) LoadCAFromPersistor() error {
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.isCALoaded() {
		return fmt.Errorf("CA is already loaded")
	}

	ca, err := m.persistor.Load()
	if err != nil {
		return err
	}
	m.ca = ca
	return nil
}

// ValidateCA validates loaded CA. Will return nil if validation passes, or the error if it fails.
// If no CA is loaded, it will also produce an error.
func (m *Manager) ValidateCA() error {
	m.lock.Lock()
	defer m.lock.Unlock()
	if !m.isCALoaded() {
		return fmt.Errorf("no CA loaded")
	}

	return m.ca.Validate()
}

// GetCA returns a deep copy of the loaded CA. Will return an error if no CA is loaded.
func (m *Manager) GetCA() (*ca.CertificateAuthority, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if !m.isCALoaded() {
		return nil, fmt.Errorf("no CA loaded")
	}

	newKey := make([]byte, len(m.ca.Key))
	copy(newKey, m.ca.Key)
	newCert := make([]byte, len(m.ca.Cert))
	copy(newCert, m.ca.Cert)
	return &ca.CertificateAuthority{
		Key:  newKey,
		Cert: newCert,
		Pass: m.ca.Pass,
	}, nil
}

// GenerateCA tries to generate a CA. It returns an error if this operation fails
// or a CA is already loaded. *NOTE:* This will **not** persist the CA, but only
// make it available to the manager!
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

// HasPersistedCA determines if the persistor interface has a CA stored.
func (m *Manager) HasPersistedCA() bool {
	return m.persistor.Exists()
}

// PersistCA tries to store the CA to the storage backed by the persistor interface.
// It returns an error if this operation fails or if no CA is loaded. If force is set,
// any existing CA at the storage layer will be overwritten. If force is not set, an
// error will be returned if there is a CA already stored.
func (m *Manager) PersistCA(force bool) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	if !m.isCALoaded() {
		return fmt.Errorf("no CA loaded")
	}

	if force {
		return m.persistor.Overwrite(*m.ca)
	}
	return m.persistor.Store(*m.ca)
}

// DeleteCA tries to delete the CA from the storage backed by the persistor interface.
// It returns an error if this operation fails, or if there is no CA stored with this
// persistor. *NOTE:* This will not unload a CA from the manager!
func (m *Manager) DeleteCA() error {
	return m.persistor.Delete()
}

// isCALoaded is the internal version of IsCALoaded but just without the lock
func (m *Manager) isCALoaded() bool {
	return m.ca != nil
}
