// Package kubernetes provides all implementations of the `persistor.Interface` that are
// done in Kubernetes. Currently there is one implementation: the `SecretsPersistor`. However, there
// are also other candidates possible - e.g. a persistor implementation which stores the CA on a
// PersistentVolume through a PersistentVolumeClaim would be another feasible implementation.
//
// The `SecretsPersistor` implements the `persistor.Interface` to store/load a CA from/to a
// Kubernetes secret. It can be used together with the CA manager in the `ca.Manager` in the `mgr`
// package to manage a CA in a Kubernetes Secret. This is currently the default implementation for
// Trireme-Kubernetes.
//
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

// SecretsPersistor implements the CA `persistor.Interface`
// and persists a CA to a Kubernetes Secret
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

// Store the given CA. It will throw the underlying Kubernetes client error if something goes wrong.
func (p *SecretsPersistor) Store(ca ca.CertificateAuthority) error {
	newSecret := caToSecret(ca, p.name, p.namespace)
	_, err := p.client.CoreV1().Secrets(p.namespace).Create(&newSecret)
	return err
}

// Delete the stored CA. It will throw the underlying Kubernetes client error if something goes wrong.
func (p *SecretsPersistor) Delete() error {
	return p.client.CoreV1().Secrets(p.namespace).Delete(p.name, &metav1.DeleteOptions{})
}

// Overwrite the stored CA. It will throw the underlying Kubernetes client error if something goes wrong.
func (p *SecretsPersistor) Overwrite(ca ca.CertificateAuthority) error {
	newSecret := caToSecret(ca, p.name, p.namespace)
	_, err := p.client.CoreV1().Secrets(p.namespace).Update(&newSecret)
	return err
}

// Exists returns true if there is a CA stored. **NOTE:** this implementation currently tries to load
// the CA from Kubernetes. If there is an error loading it, this will denote that the CA does not
// exist in the store.
func (p *SecretsPersistor) Exists() bool {
	_, err := p.client.CoreV1().Secrets(p.namespace).Get(p.name, metav1.GetOptions{})
	return err == nil
}

// Load the stored CA. It will throw the underlying Kubernetes client error if something goes wrong. It
// will furthermore throw an error if the secret cannot be converted to a `CertificateAuthority` struct.
func (p *SecretsPersistor) Load() (*ca.CertificateAuthority, error) {
	obj, err := p.client.CoreV1().Secrets(p.namespace).Get(p.name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return secretToCA(obj)
}
