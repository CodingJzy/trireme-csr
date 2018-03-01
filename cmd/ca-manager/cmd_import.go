package main

import "fmt"

// Import an existing CA
func (a *app) Import() error {
	var err error
	err = a.mgr.LoadCAFromFiles(a.config.Commands.Import.Cert, a.config.Commands.Import.Key, a.config.Commands.Import.Password)
	if err != nil {
		return fmt.Errorf("failed to load CA from files: %s", err.Error())
	}
	return nil
}
