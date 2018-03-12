package main

import "fmt"

// Import an existing CA
func (a *app) Import() error {
	var err error

	// if there is a persisted CA, and we don't force it, abort here
	if a.mgr.HasPersistedCA() && !a.config.Commands.Import.Force {
		return fmt.Errorf("persisted CA found")
	}

	// load the CA into the manager into memory
	err = a.mgr.LoadCAFromFiles(a.config.Commands.Import.Cert, a.config.Commands.Import.Key, a.config.Commands.Import.Password)
	if err != nil {
		return fmt.Errorf("failed to load CA from files: %s", err.Error())
	}

	// and store it to the persistor
	err = a.mgr.PersistCA(a.config.Commands.Import.Force)
	if err != nil {
		return fmt.Errorf("failed to persist CA: %s", err.Error())
	}

	return nil
}
