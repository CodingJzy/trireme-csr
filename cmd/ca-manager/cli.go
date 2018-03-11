package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	camgr "github.com/aporeto-inc/trireme-csr/ca/mgr"
	"github.com/aporeto-inc/trireme-csr/ca/persistor"
	kubepersistor "github.com/aporeto-inc/trireme-csr/ca/persistor/kubernetes"
)

// app holds state for all the sub-commands of the application
type app struct {
	cmd    *cobra.Command
	config *Configuration
	mgr    *camgr.Manager
}

// Execute calls the underlying cobra execute command
func (a *app) Execute() error {
	return a.cmd.Execute()
}

// initApp initializes the app and returns a reference to it
func initApp() *app {
	var app app
	var config Configuration
	// initialize viper first
	// 1. initialize our default values
	viper.SetDefault("log_level", "info")
	viper.SetDefault("log_format", "simple")
	viper.SetDefault("kube_config_path", os.Getenv("HOME")+DefaultKubeConfigLocation)
	viper.SetDefault("persistor.type", KubernetesSecretsPersistorType)
	viper.SetDefault("persistor.kubernetes_secrets.name", kubepersistor.DefaultCertificateAuthorityName)
	viper.SetDefault("persistor.kubernetes_secrets.namespace", kubepersistor.DefaultCertificateAuthorityNamespace)
	viper.SetDefault("persistor.commands.show.cert", true)
	viper.SetDefault("persistor.commands.show.key", true)
	viper.SetDefault("persistor.commands.show.key_password", false)
	viper.SetDefault("persistor.commands.generate.force", false)
	viper.SetDefault("persistor.commands.import.key", "")
	viper.SetDefault("persistor.commands.import.cert", "")
	viper.SetDefault("persistor.commands.import.password", "")
	viper.SetDefault("persistor.commands.export.key", "")
	viper.SetDefault("persistor.commands.export.cert", "")
	viper.SetDefault("persistor.commands.export.encrypt_key", true)
	viper.SetDefault("persistor.commands.export.password", "")

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
			return app.Show()
		},
	}
	cmdShow.Flags().BoolP("cert", "c", true, "Prints the CA certificate")
	cmdShow.Flags().BoolP("key", "k", true, "Prints the CA key")
	cmdShow.Flags().BoolP("key-password", "p", false, "Prints the password for the CA key")
	viper.BindPFlag("commands.show.cert", cmdShow.Flags().Lookup("cert"))
	viper.BindPFlag("commands.show.key", cmdShow.Flags().Lookup("key"))
	viper.BindPFlag("commands.show.key_password", cmdShow.Flags().Lookup("key-password"))

	// generate - generates a new CA and stores it
	cmdGenerate := &cobra.Command{
		Use:   "generate",
		Short: "Generates a new CA",
		Long:  "This will generate a new CA and store it in the defined persistor.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// execute the actual command
			return app.Generate()
		},
	}
	cmdGenerate.Flags().BoolP("force", "f", false, "Overwrites an already existing CA")
	viper.BindPFlag("Commands.Generate.Force", cmdGenerate.Flags().Lookup("force"))

	// import - imports an existing CA and stores it
	cmdImport := &cobra.Command{
		Use:   "import",
		Short: "Imports an existing CA",
		Long:  "This will import an already existing CA and store it in the defined persistor.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// execute the actual command
			return app.Import()
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
			return app.Export()
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

	cmdDelete := &cobra.Command{
		Use:   "delete",
		Short: "Deletes the existing CA",
		Long:  "This will delete an existing CA from the peristor.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// execute the actual command
			return app.Delete()
		},
	}
	cmdDelete.Flags().BoolP("force", "f", false, "Forces the deletion without confirmation")
	viper.BindPFlag("commands.delete.force", cmdDelete.Flags().Lookup("force"))

	// last but not least: the root command
	rootCmd := &cobra.Command{
		Use:   os.Args[0],
		Short: "CA Manager",
		Long:  "Command for managing the Trireme-CSR CA",
		Args:  cobra.NoArgs,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// for all commands we want to apply our viper configuration first
			err := viper.Unmarshal(&config)
			if err != nil {
				return fmt.Errorf("failed to initialize config: %s", err.Error())
			}

			// set the config in our app
			app.config = &config

			// setup logs
			err = setLogs(config.LogFormat, config.LogLevel)
			if err != nil {
				return fmt.Errorf("error setting up logs: %s", err)
			}

			zap.L().Debug("Loaded configuration", zap.Any("config", config))

			// initialize the persistor
			var persistor persistor.Interface
			switch app.config.Persistor.Type {
			case KubernetesSecretsPersistorType:
				// the Kubernetes Secrets persistor needs a kubeclient
				kubeclient, err := newKubeClient()
				if err != nil {
					zap.L().Fatal("failed to create kubernetes client", zap.Error(err))
				}

				// now initialize the kubernetes Secrets persistor
				persistor = kubepersistor.NewSecretsPersistor(
					kubeclient,
					config.Persistor.KubernetesSecrets.Name,
					config.Persistor.KubernetesSecrets.Namespace,
				)

			default:
				return fmt.Errorf("unsupported persistor type: '%s'", app.config.Persistor.Type)
			}

			// initialize the CA manager now with the configured persistor
			app.mgr, err = camgr.NewManager(persistor)
			if err != nil {
				return fmt.Errorf("failed to initialize CA Manager: %s", err.Error())
			}

			return nil
		},
	}
	rootCmd.AddCommand(cmdShow, cmdGenerate, cmdImport, cmdExport, cmdDelete)
	rootCmd.PersistentFlags().String("log-level", "info", "Log Level")
	rootCmd.PersistentFlags().String("log-format", "simple", "Log Format")
	rootCmd.PersistentFlags().String("kube-config-path", os.Getenv("HOME")+DefaultKubeConfigLocation, "Path to KubeConfig. If not found or the file does not exist, in-cluster configuration is assumed.")
	rootCmd.PersistentFlags().String("persistor-type", string(KubernetesSecretsPersistorType), "Set the persistor type. Currently only 'kubernetes-secrets' is supported.")
	rootCmd.PersistentFlags().String("persistor-kubernetes-secrets-name", kubepersistor.DefaultCertificateAuthorityName, "Sets the Kubernetes Secret name where the CA is going to be persisted to.")
	rootCmd.PersistentFlags().String("persistor-kubernetes-secrets-namespace", kubepersistor.DefaultCertificateAuthorityNamespace, "Sets the namespace where the Kubernetes Secret will be stored under.")
	viper.BindPFlag("log_level", rootCmd.PersistentFlags().Lookup("log-level"))
	viper.BindPFlag("log_format", rootCmd.PersistentFlags().Lookup("log-format"))
	viper.BindPFlag("kube_config_path", rootCmd.PersistentFlags().Lookup("kube-config-path"))
	viper.BindPFlag("persistor.type", rootCmd.PersistentFlags().Lookup("persistor-type"))
	viper.BindPFlag("persistor.kubernetes_secrets.name", rootCmd.PersistentFlags().Lookup("persistor-kubernetes-secrets-name"))
	viper.BindPFlag("persistor.kubernetes_secrets.namespace", rootCmd.PersistentFlags().Lookup("persistor-kubernetes-secrets-namespace"))

	// last but not least, set the root command in the app
	app.cmd = rootCmd

	return &app
}
