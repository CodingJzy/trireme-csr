package main

const (
	// CAManagerEnvPrefix is the environment variable prefix
	CAManagerEnvPrefix = "CAMGR"

	// DefaultKubeConfigLocation is the default location of the KubeConfig file.
	DefaultKubeConfigLocation = "/.kube/config"
)

// PersistorType type
type PersistorType string

const (
	// KubernetesSecretsPersistorType definition
	KubernetesSecretsPersistorType PersistorType = "kubernetes-secrets"
)

// Configuration is getting filled from viper with all the options
type Configuration struct {
	LogFormat      string           `mapstructure:"log_format"`
	LogLevel       string           `mapstructure:"log_level"`
	KubeConfigPath string           `mapstructure:"kube_config_path"`
	Persistor      *PersistorConfig `mapstructure:"persistor"`
	Commands       *CommandsConfig  `mapstructure:"commands"`
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
