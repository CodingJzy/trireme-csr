package main

import (
	"os"

	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	// load all client auth plugins
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

func newKubeClient() (*kubernetes.Clientset, error) {
	var err error
	var config *rest.Config
	var clientset *kubernetes.Clientset

	kubeconfig := os.Getenv("HOME") + "/.kube/config"
	_, err = os.Stat(kubeconfig)
	if err != nil && os.IsNotExist(err) {
		zap.L().Debug("trying to use InClusterConfig()")
		// try using cluster-internal config
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	} else {
		zap.L().Debug("trying to use out-of-the cluster configuration", zap.String("kubeconfig", kubeconfig))

		// try using provided config for cluster-external config
		// uses the current context in kubeconfig
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
	}

	// creates the clientset
	zap.L().Debug("trying to create clientset now from config")
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}
