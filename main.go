package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aporeto-inc/trireme-csr/certificates"
	"github.com/aporeto-inc/trireme-csr/config"

	certificateclient "github.com/aporeto-inc/trireme-csr/client"
	certificatecontroller "github.com/aporeto-inc/trireme-csr/controller"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	config, err := config.LoadConfig()
	if err != nil {
		panic("Error generating config: " + err.Error())
	}
	setLogs(config.LogFormat, config.LogLevel)

	issuer, err := certificates.NewTriremeIssuerFromPath(config.SigningCACert, config.SigningCACertKey, config.SigningCACertKeyPass)
	if err != nil {
		panic("Error creating Certificate Issuer " + err.Error())
	}

	// Get the Kube API interface for Certificates up
	kubeconfig, err := buildConfig(config.KubeconfigPath)
	if err != nil {
		panic("Error generating Kubeconfig " + err.Error())
	}

	certClient, _, err := certificateclient.NewClient(kubeconfig)
	if err != nil {
		panic("Error creating REST Kube Client for certificates: " + err.Error())
	}

	// start a controller on instances of the Certificates custom resource
	certController, err := certificatecontroller.NewCertificateController(certClient, issuer)
	if err != nil {
		panic("Error creating CertificateController" + err.Error())
	}

	go certController.Run()

	waitForSig()

	zap.L().Info("Trireme-CSR exiting")
}

// setLogs setups Zap to the specified logLevel.
func setLogs(format, logLevel string) error {
	var zapConfig zap.Config

	switch format {
	case "json":
		zapConfig = zap.NewProductionConfig()
		zapConfig.DisableStacktrace = true
	default:
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.DisableStacktrace = true
		zapConfig.DisableCaller = true
		zapConfig.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {}
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

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

func buildConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}

func waitForSig() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	zap.L().Info("Everything started. Waiting for Stop signal")
	// Waiting for a Sig
	<-c
	zap.L().Info("SIG received. Exiting")
}
