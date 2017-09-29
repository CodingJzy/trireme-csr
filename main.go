package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	MainClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic("Error creating REST Kube Client: ")
	}

	list, _ := MainClient.CoreV1().Pods("").List(metav1.ListOptions{})
	for _, i := range list.Items {
		fmt.Println(i.Name)
	}

	CertClient, _, err := certificateclient.NewClient(config)
	if err != nil {
		panic("Error creating REST Kube Client: ")
	}

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
