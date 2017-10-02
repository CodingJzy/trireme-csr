package certificates

import (
	"crypto"
	"crypto/x509"

	"github.com/aporeto-inc/tg/tglib"
)

type Signer struct {
	signingCert    *x509.Certificate
	signingKey     crypto.PrivateKey
	signingKeyPass string
}

func newSigner(signingCertPath, signingCertKeyPath, signingKeyPass string) (*Signer, error) {
	signingCert, signingKey, err = tglib.ReadCertificatePEM(signingCertPath, signingCertKeyPath, signingKeyPass)
	if err != nil {
		return nil, err
	}

	return &Signer{
		signingCert:    signingCert,
		signingKey:     signingKey,
		signingKeyPass: signingKeyPass,
	}, nil
}
