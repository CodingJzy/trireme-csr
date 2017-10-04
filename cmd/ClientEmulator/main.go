package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/aporeto-inc/trireme-csr/certificates"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	certificateclient "github.com/aporeto-inc/trireme-csr/client"
)

// KubeconfigPath is the static path to my KubeConfig
// TODO: remove
const KubeconfigPath = "/Users/bvandewa/.kube/config"

func main() {
	setLogs("info")

	config, err := buildConfig(KubeconfigPath)
	if err != nil {
		panic(err)
	}

	certClient, _, err := certificateclient.NewClient(config)
	if err != nil {
		panic("Error creating REST Kube Client: ")
	}

	certManager, err := certificates.NewCertManager("abcd", certClient)
	if err != nil {
		fmt.Printf("Error creating the cert manager")
	}

	err = certManager.GeneratePrivateKey()
	if err != nil {
		fmt.Printf("Error generating privateKey %s", err)
	}

	err = certManager.GenerateCSR()
	if err != nil {
		fmt.Printf("Error generating CSR %s", err)
	}

	err = certManager.SendAndWaitforCert(time.Minute)
	if err != nil {
		fmt.Printf("Error Sending and waiting %s", err)
	}

	cert, err := certManager.GetCert()
	if err != nil {
		fmt.Printf("Error Getting cert %s", err)
	}

	fmt.Printf("Received certificate: %+v ", cert)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	zap.L().Info("Everything started. Waiting for Stop signal")
	// Waiting for a Sig
	<-c
	zap.L().Info("SIG received. Exiting")
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

func buildConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}
