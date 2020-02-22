package main

import (
	"github.com/CodingJzy/trireme-csr/certificates"
	"github.com/CodingJzy/trireme-csr/config"
	"os"
	"os/signal"
	"syscall"
	"time"

	certificatecontroller "github.com/CodingJzy/trireme-csr/controller"
	certificateclient "github.com/CodingJzy/trireme-csr/pkg/client/clientset/versioned"
	certificateinformers "github.com/CodingJzy/trireme-csr/pkg/client/informers/externalversions"

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

	// creating OS signal handlers for shutdown handling
	sigsCh := createSignalChannel()

	issuer, err := certificates.NewTriremeIssuerFromPath(config.SigningCACert, config.SigningCACertKey, config.SigningCACertKeyPass)
	if err != nil {
		panic("Error creating Certificate Issuer " + err.Error())
	}

	// Get the Kube API interface for Certificates up
	kubeconfig, err := buildConfig(config.KubeconfigPath)
	if err != nil {
		panic("Error generating Kubeconfig " + err.Error())
	}

	// create CertificateClient
	certClient, err := certificateclient.NewForConfig(kubeconfig)
	if err != nil {
		zap.L().Fatal("Error creating CertificateClient", zap.Error(err))
	}

	// create CertificateInformer Factory for a shared informer
	certInformerFactory := certificateinformers.NewSharedInformerFactory(certClient, time.Second*30)

	// create our controller
	certController := certificatecontroller.NewCertificateController(certClient, certInformerFactory, issuer)

	// start the shared informer (internally, it calls Run(sigsCh) on the shared informer)
	certInformerFactory.Start(sigsCh)

	// start and block
	err = certController.Run(sigsCh)
	if err != nil {
		zap.L().Fatal("Error running CertificateController", zap.Error(err))
	}

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

func createSignalChannel() <-chan struct{} {
	stopCh := make(chan struct{})
	sigsCh := make(chan os.Signal, 2)
	signal.Notify(sigsCh, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	go func() {
		// wait for OS signals here,
		// close when we receive a signal
		sig := <-sigsCh
		zap.L().Info("SIG received. Exiting...", zap.String("signal", sig.String()))
		close(stopCh)

		// we will wait a second time,
		// if an immediate signal comes
		// before everything could have been shutdown
		// we force an exit
		sig = <-sigsCh
		zap.L().Info("2nd SIG received. Forcing Exit!", zap.String("signal", sig.String()))
		os.Exit(1)
	}()

	return stopCh
}
