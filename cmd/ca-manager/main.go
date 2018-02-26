package main

import (
	"go.uber.org/zap"
)

func main() {
	app := initApp()
	err := app.Execute()
	if err != nil {
		zap.L().Fatal("failed to execute command", zap.Error(err))
	}
}
