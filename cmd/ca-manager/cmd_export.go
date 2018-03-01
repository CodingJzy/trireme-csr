package main

import (
	"encoding/pem"
	"fmt"
	"io/ioutil"

	"github.com/aporeto-inc/tg/tglib"
)

// Export the installed CA
func (a *app) Export() error {
	var err error

	// if there is no CA persisted, abort immediately
	if !a.mgr.HasPersistedCA() {
		return fmt.Errorf("no persisted CA found")
	}

	// load the CA from the persistor
	err = a.mgr.LoadCAFromPersistor()
	if err != nil {
		return err
	}

	// validate the CA
	err = a.mgr.ValidateCA()
	if err != nil {
		return err
	}

	// retrieve it
	ca, err := a.mgr.GetCA()
	if err != nil {
		return err
	}

	// now write the cert to disk
	if len(a.config.Commands.Export.Cert) > 0 {
		err = ioutil.WriteFile(a.config.Commands.Export.Cert, ca.Cert, 0644)
		if err != nil {
			return fmt.Errorf("failed to write certificate to file: %s", err.Error())
		}
	}

	// check if the key needs to be decrypted, and decrypt it
	keyBytes := ca.Key
	passwordStr := ca.Pass
	if !a.config.Commands.Export.EncryptKey {
		// we don't need to write the password at all if we decrypt the key
		passwordStr = ""

		decryptedKeyPem, err := tglib.DecryptPrivateKeyPEM(ca.Key, ca.Pass)
		if err != nil {
			return fmt.Errorf("failed to decrypt private key: %s", err.Error())
		}

		keyBytes = pem.EncodeToMemory(decryptedKeyPem)
		return nil
	}

	// now write the key to disk
	if len(a.config.Commands.Export.Key) > 0 {
		err = ioutil.WriteFile(a.config.Commands.Export.Key, keyBytes, 0400)
		if err != nil {
			return fmt.Errorf("failed to write private key to file: %s", err.Error())
		}
	}

	// and last but not least write the password for the key to disk
	if len(a.config.Commands.Export.Key) > 0 && len(passwordStr) > 0 {
		err = ioutil.WriteFile(a.config.Commands.Export.Password, []byte(passwordStr), 0400)
		if err != nil {
			return fmt.Errorf("failed to write password for key to file: %s", err.Error())
		}
	}
	return nil
}
