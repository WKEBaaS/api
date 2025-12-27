// Package config BaaS API Config
package config

import (
	"bytes"
	_ "embed"
	"log"
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

//go:embed config.yaml
var defaultConfig []byte

func NewConfig(i do.Injector) (*Config, error) {
	c := &Config{}

	viper.SetConfigType("yaml")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadConfig(bytes.NewBuffer(defaultConfig)); err != nil {
		return nil, err
	}

	// 如果有外部檔案就讀取，沒有就用嵌入的 defaults
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	_ = viper.MergeInConfig()

	err := viper.Unmarshal(&c, viper.DecodeHook(
		mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToURLHookFunc(),
		),
	))
	if err != nil {
		return nil, err
	}

	log.Printf("Loaded configuration: %+v\n", c)

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
