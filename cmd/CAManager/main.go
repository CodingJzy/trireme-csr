package main

import (
	"go.uber.org/zap"

	camgr "github.com/aporeto-inc/trireme-csr/ca/mgr"
	kubepersistor "github.com/aporeto-inc/trireme-csr/ca/persistor/kubernetes"
)

func main() {
	setLogs("info")

	kubeclient, err := newKubeClient()
	if err != nil {
		zap.L().Fatal("failed to create kubernetes client", zap.Error(err))
	}

	persistor := kubepersistor.NewSecretsPersistor(
		kubeclient,
		kubepersistor.DefaultCertificateAuthorityName,
		kubepersistor.DefaultCertificateAuthorityNamespace,
	)

	mgr, err := camgr.NewManager(persistor)
	if err != nil {
		zap.L().Fatal("failed to create CA Manager", zap.Error(err))
	}

	mgr.GenerateCA()
}

// setLogs setups Zap to the specified logLevel.
func setLogs(logLevel string) error {
	zapConfig := zap.NewDevelopmentConfig()
	zapConfig.DisableStacktrace = true

	// Set the logger
	switch logLevel {
	case "trace":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "debug":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	case "fatal":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.FatalLevel)
	default:
		zapConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	logger, err := zapConfig.Build()
	if err != nil {
		return err
	}

	zap.ReplaceGlobals(logger)
	return nil
}
