// Package ca holds everything that is related to the CA for Trireme-CSR. In the ca package itself
// is just one high-level struct called `CertificateAuthority` - it holds a CA in memory. This struct is
// heavily used inside of the CA Manager (in the `mgr` sub-package) which offers functionality around
// managing a CA. The `persistor` sub-package defines a high-level layer on how to store/load a CA to a
// storage of some kind.
package ca

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/rand"
	"time"

	"github.com/aporeto-inc/tg/tglib"
)

const (
	// passwordLength is the length of a password when it is auto-generated for a CA
	passwordLength = 32
)

// CertificateAuthority holds a CA. This is still in its infancies, however, it works very well
// like this together with tglib.
type CertificateAuthority struct {
	Key  []byte
	Pass string
	Cert []byte
}

// LoadCertificateAuthorityFromFiles loads an existing CA from files
func LoadCertificateAuthorityFromFiles(certPath, keyPath, pass string) (*CertificateAuthority, error) {
	var err error
	var cert, key []byte
	cert, err = ioutil.ReadFile(certPath)
	if err != nil {
		return nil, err
	}
	key, err = ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}
	ca := &CertificateAuthority{
		Key:  key,
		Cert: cert,
		Pass: pass,
	}
	err = ca.Validate()
	if err != nil {
		return nil, err
	}
	return ca, nil
}

// NewCertificateAuthority generates a new CA
func NewCertificateAuthority() (*CertificateAuthority, error) {
	// generate a random password of `passwordLength` characters which consists of
	// alphanumeric characters plus a selected set of special characters
	password := randomPassword()

	certPem, keyPem, err := tglib.IssueCertiticate(nil, nil,
		tglib.ECPrivateKeyGenerator, nil, nil, nil, nil, nil, nil, nil,
		"Trireme-CSR CA", nil, nil,
		time.Now(), time.Now().Add(365*24*time.Hour), x509.KeyUsageCRLSign|x509.KeyUsageCertSign, nil,
		x509.ECDSAWithSHA384, x509.ECDSA, true, nil,
	)
	if err != nil {
		return nil, fmt.Errorf("tglib.IssueCertificate: %s", err.Error())
	}
	keyPemEncrypted, err := tglib.EncryptPrivateKey(keyPem, password)
	if err != nil {
		return nil, fmt.Errorf("tglib.EncryptPrivateKey: %s", err.Error())
	}
	ca := &CertificateAuthority{
		Key:  pem.EncodeToMemory(keyPemEncrypted),
		Cert: pem.EncodeToMemory(certPem),
		Pass: password,
	}
	err = ca.Validate()
	if err != nil {
		return nil, fmt.Errorf("ca.Validate: %s", err.Error())
	}
	return ca, nil
}

// Validate the CA. Will return nil if validation passes, or the error if it fails.
func (ca *CertificateAuthority) Validate() error {
	_, _, err := tglib.ReadCertificate(ca.Cert, ca.Key, ca.Pass)
	return err
}

// randomPassword provides a random password for the `NewCertificateAuthority` function
func randomPassword() string {
	// tribute goes to https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-golang
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_,.<>/?:;{}[]+"
	const (
		letterIdxBits = 6                    // 6 bits to represent a letter index
		letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
		letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
	)
	n := passwordLength
	src := rand.NewSource(time.Now().UnixNano())
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}
