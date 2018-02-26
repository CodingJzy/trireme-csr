package kubernetes

import (
	"github.com/aporeto-inc/trireme-csr/ca"
	"github.com/aporeto-inc/trireme-csr/ca/persistor"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	// DefaultCertificateAuthorityName is a good fallback for the secret name
	DefaultCertificateAuthorityName = "trireme-cacert"
	// DefaultCertificateAuthorityNamespace is the kubernetes namespace where this secret should be stored
	DefaultCertificateAuthorityNamespace = "kube-system"
)

// SecretsPersistor implements the CA persistor interface
// and persists a CA to a Kubernetes secret
type SecretsPersistor struct {
	client    *kubernetes.Clientset
	name      string
	namespace string
}

// NewSecretsPersistor creates a new CA Kubernetes Secrets Persistor
func NewSecretsPersistor(client *kubernetes.Clientset, name, namespace string) persistor.Interface {
	return &SecretsPersistor{
		client:    client,
		name:      name,
		namespace: namespace,
	}
}

// Store the given CA
func (p *SecretsPersistor) Store(ca ca.CertificateAuthority) error {
	newSecret := caToSecret(ca, p.name, p.namespace)
	_, err := p.client.CoreV1().Secrets(p.namespace).Create(&newSecret)
	return err
}

// Delete the stored CA
func (p *SecretsPersistor) Delete() error {
	return p.client.CoreV1().Secrets(p.namespace).Delete(p.name, &metav1.DeleteOptions{})
}

// Overwrite the stored CA
func (p *SecretsPersistor) Overwrite(ca ca.CertificateAuthority) error {
	newSecret := caToSecret(ca, p.name, p.namespace)
	_, err := p.client.CoreV1().Secrets(p.namespace).Update(&newSecret)
	return err
}

// Exists returns true if there is a CA stored
func (p *SecretsPersistor) Exists() bool {
	_, err := p.client.CoreV1().Secrets(p.namespace).Get(p.name, metav1.GetOptions{})
	return err == nil
}

// Load the stored CA
func (p *SecretsPersistor) Load() (*ca.CertificateAuthority, error) {
	obj, err := p.client.CoreV1().Secrets(p.namespace).Get(p.name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return secretToCA(obj)
}
