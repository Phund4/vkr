package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

var (
	ErrLoadMetricsConfig    = errors.New("failed to load metrics config")
	ErrMissingEnvVariable   = errors.New("missing environment variable")
	ErrLoadServerConfig     = errors.New("failed to load server config")
	errLoadClickhouseConfig = errors.New("failed to load clickhouse config")
	ErrEnvVarFromFile       = errors.New("couldn't get environment variables from file")
)

const (
	EnvKey = "ENV"
)

// Config содержит всю конфигурацию приложения
type Config struct {
	Metrics    MetricsConfig
	Env        string
	Server     ServerConfig
	Clickhouse ClickhouseConfig
}

// LoadConfig загружает всю конфигурацию из переменных окружения
func LoadConfig() (*Config, error) {
	env := os.Getenv(EnvKey)
	if env == "" {
		return nil, fmt.Errorf("%w: %s", ErrMissingEnvVariable, EnvKey)
	}

	serverCfg, err := loadServerConfig()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrLoadServerConfig, err)
	}

	metricsCfg, err := loadMetricsConfig()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrLoadMetricsConfig, err)
	}

	clickhouseCfg, err := loadClickhouseConfig()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errLoadClickhouseConfig, err)
	}

	return &Config{
		Env:        env,
		Server:     serverCfg,
		Metrics:    metricsCfg,
		Clickhouse: clickhouseCfg,
	}, nil
}

// SetupViper настраивает Viper для чтения переменных окружения.
func SetupViper(configFile, configType string) error {
	if err := godotenv.Load(configFile); err != nil {
		return fmt.Errorf("%w: %w", ErrEnvVarFromFile, err)
	}

	viper.SetConfigFile(configFile)
	viper.SetConfigType(configType)
	viper.AutomaticEnv()

	return nil
}

// LoadAdditionalEnv загружает дополнительный .env файл поверх уже загруженных
func LoadAdditionalEnv(configFile string) error {
	if err := godotenv.Load(configFile); err != nil {
		return fmt.Errorf("%w: %w", ErrEnvVarFromFile, err)
	}
	return nil
}
