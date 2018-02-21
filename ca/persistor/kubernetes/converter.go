package kubernetes

import (
	"github.com/aporeto-inc/trireme-csr/ca"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	secretCertEntry = "ca-cert.pem"
	secretKeyEntry  = "ca-key.pem"
	secretPwEntry   = "ca-pass"
)

func caToSecret(ca ca.CertificateAuthority, name, namespace string) corev1.Secret {
	return corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			secretKeyEntry:  ca.Key,
			secretPwEntry:   []byte(ca.Pass),
			secretCertEntry: ca.Cert,
		},
	}
}
