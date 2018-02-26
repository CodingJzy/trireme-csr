package main

import (
	"fmt"
)

// Show prints the CA to the screen
func (a *app) Show() error {
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

	if a.config.Commands.Show.Cert {
		fmt.Println("CA Certificate:")
		fmt.Println(string(ca.Cert))
		fmt.Println()
	}

	if a.config.Commands.Show.Key {
		fmt.Println("CA Key:")
		fmt.Println(string(ca.Key))
		fmt.Println()
	}

	if a.config.Commands.Show.KeyPassword {
		fmt.Println("CA Key Password:")
		fmt.Println(ca.Pass)
		fmt.Println()
	}
	return nil
}
