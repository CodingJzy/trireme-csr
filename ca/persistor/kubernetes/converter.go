package kubernetes

import (
	"fmt"

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

func secretToCA(secret *corev1.Secret) (*ca.CertificateAuthority, error) {
	key, ok := secret.Data[secretKeyEntry]
	if !ok {
		return nil, fmt.Errorf("entry not in secret: '%s'", secretKeyEntry)
	}
	cert, ok := secret.Data[secretCertEntry]
	if !ok {
		return nil, fmt.Errorf("entry not in secret: '%s'", secretCertEntry)
	}
	password, ok := secret.Data[secretPwEntry]
	if !ok {
		return nil, fmt.Errorf("entry not in secret: '%s'", secretPwEntry)
	}

	return &ca.CertificateAuthority{
		Key:  key,
		Pass: string(password),
		Cert: cert,
	}, nil
}
