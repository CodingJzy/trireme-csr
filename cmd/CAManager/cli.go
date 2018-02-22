package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
			Name:      "trireme-cacert",
			Namespace: "kube-system",
		},
	})
	viper.SetDefault("commands", &CommandsConfig{
		Show: &ShowCmdConfig{
			Cert:        true,
			Key:         true,
			KeyPassword: false,
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
	// export - exports an existing CA to files if it exists

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
	rootCmd.AddCommand(cmdShow)
	rootCmd.PersistentFlags().String("log-level", "info", "Log level")
	rootCmd.PersistentFlags().String("log-format", "", "Log Format")
	viper.BindPFlag("LogLevel", rootCmd.PersistentFlags().Lookup("log-level"))
	viper.BindPFlag("LogFormat", rootCmd.PersistentFlags().Lookup("log-format"))

	return nil
}
