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
	App         App         `yaml:"app"`
	Server      Server      `yaml:"server"`
	Persistence Persistence `yaml:"persistence"`
	JWT         JWT         `yaml:"jwt"`
	WorkerPool  WorkerPool  `yaml:"workerpool"`
	Service     Service     `yaml:"service"`
}

type App struct {
	Name    string `yaml:"name" validate:"required"`
	Version string `yaml:"version" validate:"required,semver"`
	Env     string `yaml:"env" validate:"required,oneof=dev local prod"`
}

type Server struct {
	HTTP struct {
		Addr string `yaml:"addr" validate:"required,hostname_port"`

		// There is no validation here because this field
		// is unique and will be checked separately in
		// the functions only if `enable == true`.
		TLS struct {
			Enable         bool   `yaml:"enable"`
			ServerCertPath string `yaml:"server_cert_path"`
			ServerKeyPath  string `yaml:"server_key_path"`
		}
	} `yaml:"http"`
	Conns struct {
		ReadTimeout  time.Duration `yaml:"read_timeout" validate:"required,min=100ms"`
		WriteTimeout time.Duration `yaml:"write_timeout" validate:"required,min=100ms"`
		IdleTimeout  time.Duration `yaml:"idle_timeout" validate:"required,min=100ms"`
	} `yaml:"conns"`
}

type Persistence struct {
	MigrationsPath string `yaml:"migrations_path" validate:"omitempty"`
	Postgres       struct {
		Host    string `yaml:"host" validate:"required,hostname"`
		Port    int    `yaml:"port" validate:"required,gte=1,lte=65535"`
		SSLMode string `yaml:"sslmode" validate:"required,oneof=disable enable"`
		Auth    struct {
			User     string `yaml:"user" validate:"required"`
			Password string `yaml:"password" validate:"required"`
			DBName   string `yaml:"dbname" validate:"required"`
		} `yaml:"auth"`
		Conns struct {
			MaxIdles    int           `yaml:"max_idles" validate:"required,gte=1"`
			MaxOpens    int           `yaml:"max_opens" validate:"required,gte=1"`
			MaxLifetime time.Duration `yaml:"max_lifetime" validate:"required,gte=1m"`
			MaxIdleTime time.Duration `yaml:"max_idle_time" validate:"required,gte=1m"`
		} `yaml:"conns"`
	} `yaml:"postgres"`
}

type JWT struct {
	PrivatePath string        `yaml:"private_path" validate:"required"`
	PublicPath  string        `yaml:"public_path" validate:"required"`
	TokenTTL    time.Duration `yaml:"token_ttl" validate:"required,min=1m"`
}

type WorkerPool struct {
	Workers int64 `yaml:"workers" validate:"required,gte=1"`
	TaskBuf int64 `yaml:"task_buf" validate:"required,gte=1"`
	ErrBuf  int64 `yaml:"err_buf" validate:"required,gte=1"`
}

type Service struct {
	TickerInterval    time.Duration `yaml:"ticker_interval" validate:"required,min=100ms"`
	NotificateTimeout time.Duration `yaml:"notificate_timeout" validate:"required,min=100ms"`
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
