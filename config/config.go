package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/viper"

	flag "github.com/spf13/pflag"
)

// DefaultKubeConfigLocation is the default location of the KubeConfig file.
const DefaultKubeConfigLocation = "/.kube/config"

// Configuration contains all the User Parameter for Trireme-Kubernetes.
type Configuration struct {
	// AuthType defines if Trireme uses PSK or PKI
	InstallCRD bool

	GenerateCA bool

	SigningCACert        string
	SigningCACertData    []byte
	SigningCACertKey     string
	SigningCACertKeyData []byte
	SigningCACertKeyPass string

	KubeconfigPath string
	LogLevel       string
}

func usage() {
	flag.PrintDefaults()
	os.Exit(2)
}

// LoadConfig loads a Configuration struct:
// 1) If presents flags are used
// 2) If no flags, Env Variables are used
// 3) If no Env Variables, defaults are used when possible.
func LoadConfig() (*Configuration, error) {
	flag.Usage = usage
	flag.Bool("InstallCRD", false, "Install CRD if not initialized ?")
	flag.String("KubeconfigPath", "", "KubeConfig used to connect to Kubernetes")
	flag.String("LogLevel", "", "Log level. Default to info (trace//debug//info//warn//error//fatal)")

	flag.String("SigningCacert", "", "Path to the CA that will issue certificates.")
	flag.String("signingCacertKey", "", "Path to the CA key that will issue certificates.")
	flag.String("SigningCacertPass", "", "Password for the signing CA.")

	// Setting up default configuration
	viper.SetDefault("InstallCRD", false)
	viper.SetDefault("KubeconfigPath", "")
	viper.SetDefault("LogLevel", "info")

	// Binding ENV variables
	// Each config will be of format TRIREME_XYZ as env variable, where XYZ
	// is the upper case config.
	viper.SetEnvPrefix("TRIREME")
	viper.AutomaticEnv()

	// Binding CLI flags.
	flag.Parse()
	viper.BindPFlags(flag.CommandLine)

	var config Configuration

	err := viper.Unmarshal(&config)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling:%s", err)
	}

	err = validateConfig(&config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// validateConfig is validating the Configuration struct.
func validateConfig(config *Configuration) error {
	// Validating KUBECONFIG
	// In case not running as InCluster, we try to infer a possible KubeConfig location
	if os.Getenv("KUBERNETES_PORT") == "" {
		if config.KubeconfigPath == "" {
			config.KubeconfigPath = os.Getenv("HOME") + DefaultKubeConfigLocation
		}
	} else {
		config.KubeconfigPath = ""
	}

	signingcadata, err := ioutil.ReadFile(config.SigningCACert)
	if err != nil {
		return fmt.Errorf("unable to read signing CA file: %s", err.Error())
	}

	signingcakeydata, err := ioutil.ReadFile(config.SigningCACertKey)
	if err != nil {
		return fmt.Errorf("unable to read signing CA key file: %s", err.Error())
	}

	config.SigningCACertData = signingcadata
	config.SigningCACertKeyData = signingcakeydata

	return nil
}
