package certificates

import "time"

type CertManager struct {
	keyPass    string
	privateKey []byte
	csr        []byte
}

func newCertManager() (*CertManager, error) {
	return &CertManager{}, nil
}

func (m *CertManager) GeneratePrivateKey() {
}

func (m *CertManager) GetCertificate() {
}

func (m *CertManager) GetKey() {
}

// SendAndWaitforCert is a blocking func that issue the CertificateRequest and
// returns once the Certificate is available.
func (m *CertManager) SendAndWaitforCert(timeout time.Duration) {
}

// IsIssued returns true if the certificate is issued by the controller
func (m *CertManager) IsIssued() (bool, error) {
	return false, nil
}
