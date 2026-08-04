package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/goccy/go-yaml"
)

var (
	ErrInvalidServerCert = errors.New("invalid path to server cert")
	ErrInvalidServerKey  = errors.New("invalid server key")
)

type Config struct {
	App      App      `yaml:"app"`
	Server   Server   `yaml:"server"`
	Backend  Backend  `yaml:"backend"`
	Telegram Telegram `yaml:"telegram"`
}

type App struct {
	Name    string `yaml:"name" validate:"required"`
	Version string `yaml:"version" validate:"required,semver"`
	Env     string `yaml:"env" validate:"required,oneof=dev local prod"`
}

type Server struct {
	HTTP struct {
		Addr string `yaml:"addr" validate:"required,hostname_port"`

		// There's no validation cuz its unique field,
		// validation of this will be in another function
		// if `enable == true`
		TLS struct {
			Enable         bool   `yaml:"enable"`
			ServerCertPath string `yaml:"server_cert_path"`
			ServerKeyPath  string `yaml:"server_key_path"`
		} `yaml:"tls"`
	} `yaml:"http"`
	Conns struct {
		ReadTimeout  time.Duration `yaml:"read_timeout"`
		WriteTimeout time.Duration `yaml:"write_timeout"`
		IdleTimeout  time.Duration `yaml:"idle_timeout"`
	} `yaml:"conns"`
}

type Backend struct {
	Addr    string        `yaml:"addr" validate:"required,url"`
	Timeout time.Duration `yaml:"timeout" validate:"required,min=100ms"`
}

type Telegram struct {
	Token string `yaml:"token" validate:"required"`
	Poll  struct {
		Timeout time.Duration `yaml:"timeout" validate:"required,min=100ms"`
	} `yaml:"poll"`
}

// New is a constructor for Config
//
// Calling it means that the .env file has been loaded
func New(path string) (*Config, error) {
	bytes, err := loadBytes(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load config file: %w", err)
	}

	cfg, err := parseBytes(bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	return cfg, nil
}

func validateConfig(cfg *Config) error {
	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return err
	}

	return validateTLS(cfg)
}

func validateTLS(cfg *Config) error {
	if !cfg.Server.HTTP.TLS.Enable {
		return nil
	}

	serverCertPath := filepath.Clean(cfg.Server.HTTP.TLS.ServerCertPath)
	if stat, err := os.Stat(serverCertPath); err != nil || stat.IsDir() {
		return ErrInvalidServerCert
	}

	serverKeyPath := filepath.Clean(cfg.Server.HTTP.TLS.ServerKeyPath)
	if stat, err := os.Stat(serverKeyPath); err != nil || stat.IsDir() {
		return ErrInvalidServerKey
	}

	return nil
}

func parseBytes(bytes []byte) (*Config, error) {
	var cfg Config
	if err := yaml.Unmarshal(bytes, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func loadBytes(path string) ([]byte, error) {
	path = filepath.Clean(path)
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	content := os.ExpandEnv(string(bytes))
	return []byte(content), nil
}
