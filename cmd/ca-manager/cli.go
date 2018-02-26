package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	kubepersistor "github.com/aporeto-inc/trireme-csr/ca/persistor/kubernetes"
)

const (
	// CAManagerEnvPrefix is the environment variable prefix
	CAManagerEnvPrefix = "CAMGR"
)

// PersistorType type
type PersistorType string

const (
	// KuberetesSecretsPersistorType definition
	KuberetesSecretsPersistorType PersistorType = "kubernetes-secrets"
)

// Configuration is getting filled from viper with all the options
type Configuration struct {
	LogFormat string           `mapstructure:"log_format"`
	LogLevel  string           `mapstructure:"log_level"`
	Persistor *PersistorConfig `mapstructure:"persistor"`
	Commands  *CommandsConfig  `mapstructure:"commands"`
}

// PersistorConfig struct
type PersistorConfig struct {
	Type              PersistorType            `mapstructure:"type"`
	KubernetesSecrets *KubernetesSecretsConfig `mapstructure:"kubernetes_secrets"`
}

// KubernetesSecretsConfig struct
type KubernetesSecretsConfig struct {
	Name      string `mapstructure:"name"`
	Namespace string `mapstructure:"namespace"`
}

// CommandsConfig struct
type CommandsConfig struct {
	Show     *ShowCmdConfig     `mapstructure:"show"`
	Generate *GenerateCmdConfig `mapstructure:"generate"`
	Import   *ImportCmdConfig   `mapstructure:"import"`
	Export   *ExportCmdConfig   `mapstructure:"export"`
}

// ShowCmdConfig struct
type ShowCmdConfig struct {
	Cert        bool `mapstructure:"cert"`
	Key         bool `mapstructure:"key"`
	KeyPassword bool `mapstructure:"key_password"`
}

// GenerateCmdConfig struct
type GenerateCmdConfig struct {
	Force bool `mapstructure:"force"`
}

// ImportCmdConfig struct
type ImportCmdConfig struct {
	Key      string `mapstructure:"key"`
	Cert     string `mapstructure:"cert"`
	Password string `mapstructure:"password"`
}

// ExportCmdConfig struct
type ExportCmdConfig struct {
	Key        string `mapstructure:"key"`
	Cert       string `mapstructure:"cert"`
	EncryptKey bool   `mapstructure:"encrypt_key"`
	Password   string `mapstructure:"password"`
}

// initCLI initializes the cobra CLI and returns the root command
func initCLI(showFunc, generateFunc, importFunc, exportFunc func(*Configuration) error, setLogs func(logFormat, logLevel string) error) *cobra.Command {
	var config Configuration
	// initialize viper first
	// 1. initialize our default values
	viper.SetDefault("log_level", "info")
	viper.SetDefault("log_format", "json")
	viper.SetDefault("persistor", &PersistorConfig{
		Type: KuberetesSecretsPersistorType,
		KubernetesSecrets: &KubernetesSecretsConfig{
			Name:      kubepersistor.DefaultCertificateAuthorityName,
			Namespace: kubepersistor.DefaultCertificateAuthorityNamespace,
		},
	})
	viper.SetDefault("commands", &CommandsConfig{
		Show: &ShowCmdConfig{
			Cert:        true,
			Key:         true,
			KeyPassword: false,
		},
		Generate: &GenerateCmdConfig{
			Force: false,
		},
		Import: &ImportCmdConfig{
			Key:      "",
			Cert:     "",
			Password: "",
		},
		Export: &ExportCmdConfig{
			Key:        "",
			Cert:       "",
			EncryptKey: true,
			Password:   "",
		},
	})

	// 2. read config file: first one will be taken into account
	viper.SetConfigName("ca-manager")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.trireme-csr/")
	viper.AddConfigPath("/etc/trireme-csr/")
	viper.MergeInConfig()

	// 3. setup environment variables
	viper.SetEnvPrefix(CAManagerEnvPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// show - prints the current CA if it exists
	cmdShow := &cobra.Command{
		Use:   "show [OPTIONS]",
		Short: "Prints the current CA if it exists",
		Long:  "Loads the currently existing CA into memory, and prints it to the screen.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// execute the actual command
			return showFunc(&config)
		},
	}
	cmdShow.Flags().BoolP("cert", "c", true, "Prints the CA certificate")
	cmdShow.Flags().BoolP("key", "k", true, "Prints the CA key")
	cmdShow.Flags().BoolP("key-password", "p", false, "Prints the password for the CA key")
	viper.BindPFlag("commands.show.cert", cmdShow.Flags().Lookup("cert"))
	viper.BindPFlag("commands.show.key", cmdShow.Flags().Lookup("key"))
	viper.BindPFlag("commands.show.key_password", cmdShow.Flags().Lookup("key_password"))

	// generate - generates a new CA and stores it
	cmdGenerate := &cobra.Command{
		Use:   "generate",
		Short: "Generates a new CA",
		Long:  "This will generate a new CA and store it in the defined persistor.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// execute the actual command
			return generateFunc(&config)
		},
	}
	cmdGenerate.Flags().BoolP("force", "f", false, "Overwrites an already existing CA")
	viper.BindPFlag("commands.generate.force", cmdGenerate.Flags().Lookup("force"))

	// import - imports an existing CA and stores it
	cmdImport := &cobra.Command{
		Use:   "import",
		Short: "Imports an existing CA",
		Long:  "This will import an already existing CA and store it in the defined persistor.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// execute the actual command
			return importFunc(&config)
		},
	}
	cmdImport.Flags().StringP("key", "k", "", "Path to the key file of the CA.")
	cmdImport.Flags().StringP("cert", "c", "", "Path to the certificate file of the CA.")
	cmdImport.Flags().StringP("password", "p", "", "Password to decrypt the key file of the CA.")
	viper.BindPFlag("commands.import.key", cmdImport.Flags().Lookup("key"))
	viper.BindPFlag("commands.import.cert", cmdImport.Flags().Lookup("cert"))
	viper.BindPFlag("commands.import.password", cmdImport.Flags().Lookup("password"))

	// export - exports an existing CA to files if it exists
	cmdExport := &cobra.Command{
		Use:   "export",
		Short: "Exports the existing CA",
		Long:  "This will export the currently existing CA: load it from the persistor and store it to disk.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// execute the actual command
			return exportFunc(&config)
		},
	}
	cmdExport.Flags().StringP("key", "k", "", "Path to the key file - where to write the key of the CA")
	cmdExport.Flags().BoolP("encrypt-key", "e", true, "Enables encryption of the key file")
	cmdExport.Flags().StringP("cert", "c", "", "Path to the cert file - where to write the certificate of the CA")
	cmdExport.Flags().StringP("password", "p", "", "Path to the password file - where to write the encryption key of the key. Not used if encryption is disabled. If empty and encryption is enabled, it will print the encryption key to the screen.")
	viper.BindPFlag("commands.export.key", cmdExport.Flags().Lookup("key"))
	viper.BindPFlag("commands.export.encrypt_key", cmdExport.Flags().Lookup("encrypt-key"))
	viper.BindPFlag("commands.export.cert", cmdExport.Flags().Lookup("cert"))
	viper.BindPFlag("commands.export.password", cmdExport.Flags().Lookup("password"))

	// last but not least: the root command
	rootCmd := &cobra.Command{
		Use:  "",
		Long: "Command for launching programs with Trireme policy.",
		Args: cobra.NoArgs,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// for all commands we want to apply our viper configuration first
			err := viper.Unmarshal(&config)
			if err != nil {
				return fmt.Errorf("failed to initialize config: %s", err.Error())
			}

			// setup logs
			err = setLogs(config.LogFormat, config.LogLevel)
			if err != nil {
				return fmt.Errorf("error setting up logs: %s", err)
			}
			return nil
		},
	}
	rootCmd.AddCommand(cmdShow, cmdGenerate, cmdImport, cmdExport)
	rootCmd.PersistentFlags().String("log-level", "info", "Log level")
	rootCmd.PersistentFlags().String("log-format", "", "Log Format")
	rootCmd.PersistentFlags().String("persistor-type", string(KuberetesSecretsPersistorType), "Set the persistor type. Currently only 'kubernetes-secrets' is supported.")
	rootCmd.PersistentFlags().String("persistor-kubernetes-secrets-name", kubepersistor.DefaultCertificateAuthorityName, "Sets the Kubernetes Secret name where the CA is going to be persisted to.")
	rootCmd.PersistentFlags().String("persistor-kubernetes-secrets-namespace", kubepersistor.DefaultCertificateAuthorityNamespace, "Sets the namespace where the Kubernetes Secret will be stored under.")
	viper.BindPFlag("log_level", rootCmd.PersistentFlags().Lookup("log-level"))
	viper.BindPFlag("log_format", rootCmd.PersistentFlags().Lookup("log-format"))
	viper.BindPFlag("persistor.type", rootCmd.PersistentFlags().Lookup("persistor-type"))
	viper.BindPFlag("persistor.kubernetes_secrets.name", rootCmd.PersistentFlags().Lookup("persistor-kubernetes-secrets-name"))
	viper.BindPFlag("persistor.kubernetes_secrets.namespace", rootCmd.PersistentFlags().Lookup("persistor-kubernetes-secrets-namespace"))

	return rootCmd
}
