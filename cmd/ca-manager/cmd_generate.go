package main

import "fmt"

// Generate a new CA and store it to the persistor
func (a *app) Generate() error {
	var err error

	// generate a new CA now
	err = a.mgr.GenerateCA()
	if err != nil {
		return fmt.Errorf("failed to generate CA: %s", err.Error())
	}

	// and store it to the persistor
	err = a.mgr.PersistCA(a.config.Commands.Generate.Force)
	if err != nil {
		return fmt.Errorf("failed to persist CA: %s", err.Error())
	}

	return nil
}
