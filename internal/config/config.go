// Package config BaaS API Config
package config

import (
	"log/slog"
	"net/url"
	"strings"

	// "github.com/go-viper/mapstructure/v2"
	"github.com/go-viper/mapstructure/v2"
	_ "github.com/joho/godotenv/autoload"
	"github.com/samber/do/v2"
	"github.com/spf13/viper"
)

type AppConfig struct {
	Port           string
	Host           string
	TrustedOrigins []string
	ExternalDomain string
	ExternalSecure bool
}

type DatabaseConfig struct {
	URL string
}

type AuthConfig struct {
	URL *url.URL
}

type PgRESTConfig struct {
	URL *url.URL
}

type KubeConfig struct {
	ConfigPath string
	Project    struct {
		Namespace     string
		TLSSecretName string
	}
}

type S3Config struct {
	Endpoint        string
	UseSSL          bool
	AccessKeyID     string
	SecretAccessKey string
	Region          string
}

type LoggingConfig struct {
	Level string
}

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Auth     AuthConfig
	PgREST   PgRESTConfig
	Kube     KubeConfig
	S3       S3Config
	Logging  LoggingConfig
}

func NewConfig(i do.Injector) (*Config, error) {
	c := &Config{}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("internal/config")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		return c, err
	}

	err := viper.Unmarshal(&c, viper.DecodeHook(
		mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToURLHookFunc(),
		),
	))
	if err != nil {
		return nil, err
	}

	switch c.Logging.Level {
	case "DEBUG":
		slog.SetLogLoggerLevel(slog.LevelDebug)
	case "INFO":
		slog.SetLogLoggerLevel(slog.LevelInfo)
	case "WARN":
		slog.SetLogLoggerLevel(slog.LevelWarn)
	case "ERROR":
		slog.SetLogLoggerLevel(slog.LevelError)
	default:
		slog.SetLogLoggerLevel(slog.LevelInfo)
	}

	return c, nil
}

var Package = do.Package(
	do.Lazy(NewConfig),
)
