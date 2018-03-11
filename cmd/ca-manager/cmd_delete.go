package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"
)

// Delete will try to delete the persisted CA from the selected persistor backend interface
func (a *app) Delete() error {
	var err error

	// if there is no CA persisted, abort immediately
	if !a.mgr.HasPersistedCA() {
		return fmt.Errorf("no persisted CA found")
	}

	// Get confirmation from the user
	if !a.config.Commands.Delete.Force {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Type 'yes' to delete the persisted CA: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to get user input: %s", err.Error())
		}
		if strings.ToLower(strings.TrimSpace(input)) != "yes" {
			zap.L().Error("User did not confirm the operation, aborting...")
			return fmt.Errorf("aborting operation")
		}
	}

	// now delete the CA
	err = a.mgr.DeleteCA()
	if err != nil {
		return err
	}

	return nil
}
