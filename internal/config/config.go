// Package config loads service configuration from a YAML file (mounted from a
// Kubernetes ConfigMap in a cluster) with environment-variable overrides for
// secrets. Precedence is: environment variable > YAML value > env-default tag.
package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

// PathEnv names the environment variable holding the config file location.
// When it is unset the service runs on defaults plus environment overrides,
// which is what local development without a config file relies on.
const PathEnv = "CONFIG_PATH"

type Config struct {
	App      App      `yaml:"app"`
	Postgres Postgres `yaml:"postgres"`
}

type App struct {
	// Env is the deployment environment: local, staging, production.
	Env string `yaml:"env" env:"APP_ENV" env-default:"local"`
	// ShutdownTimeout bounds graceful shutdown.
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" env:"APP_SHUTDOWN_TIMEOUT" env-default:"10s"`
}

type Postgres struct {
	Host     string `yaml:"host" env:"POSTGRES_HOST" env-default:"localhost"`
	Port     int    `yaml:"port" env:"POSTGRES_PORT" env-default:"5432"`
	User     string `yaml:"user" env:"POSTGRES_USER" env-default:"catalog"`
	Database string `yaml:"database" env:"POSTGRES_DB" env-default:"catalog"`
	SSLMode  string `yaml:"ssl_mode" env:"POSTGRES_SSLMODE" env-default:"disable"`

	// Password is deliberately not readable from YAML: it must come from a
	// Kubernetes Secret through the environment, never from a ConfigMap.
	Password string `yaml:"-" env:"POSTGRES_PASSWORD" env-default:"catalog"`

	// Pool tuning for database/sql. Defaults are chosen over the standard
	// library ones, which leave connections unlimited and idle capped at 2.
	MaxOpenConns    int           `yaml:"max_open_conns" env:"POSTGRES_MAX_OPEN_CONNS" env-default:"25"`
	MaxIdleConns    int           `yaml:"max_idle_conns" env:"POSTGRES_MAX_IDLE_CONNS" env-default:"25"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" env:"POSTGRES_CONN_MAX_LIFETIME" env-default:"30m"`
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time" env:"POSTGRES_CONN_MAX_IDLE_TIME" env-default:"5m"`
}

// DSN renders the connection string. The password is escaped rather than
// interpolated, so characters like @ or / in it cannot corrupt the URL.
func (p Postgres) DSN() string {
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(p.User, p.Password),
		Host:     net.JoinHostPort(p.Host, strconv.Itoa(p.Port)),
		Path:     "/" + p.Database,
		RawQuery: url.Values{"sslmode": {p.SSLMode}}.Encode(),
	}

	return u.String()
}

// Load reads the config file named by CONFIG_PATH, then applies environment
// overrides. With CONFIG_PATH unset only the environment and defaults are used.
func Load() (*Config, error) {
	var cfg Config

	path := os.Getenv(PathEnv)
	if path == "" {
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			return nil, fmt.Errorf("read config from environment: %w", err)
		}

		return &cfg, nil
	}

	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		return nil, fmt.Errorf("read config %q: %w", path, err)
	}

	return &cfg, nil
}
