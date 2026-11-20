package env

import (
	"errors"
	"fmt"
	"net"
	"os"

	cenv "github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	def "github.com/ossirdian/url_shortener/internal/config"
)

type rawConfig struct {
	LoggerLevel   string `env:"LOGGER_LEVEL" envDefault:"info"`
	LoggerAsJSON  bool   `env:"LOGGER_AS_JSON" envDefault:"false"`
	HTTPHost      string `env:"HTTP_HOST" envDefault:"0.0.0.0"`
	HTTPPort      string `env:"HTTP_PORT" envDefault:"8084"`
	AuthUsername  string `env:"AUTH_USERNAME,required"`
	AuthPassword  string `env:"AUTH_PASSWORD,required"`
	PostgresDSN   string `env:"POSTGRES_DSN,required"`
	MigrationsDir string `env:"MIGRATIONS_DIR" envDefault:"./migrations"`
	AutoMigrate   bool   `env:"AUTO_MIGRATE" envDefault:"false"`
}

type Config struct {
	raw rawConfig
}

type (
	loggerConfig struct {
		raw rawConfig
	}
	httpConfig struct {
		raw rawConfig
	}
	authConfig struct {
		raw rawConfig
	}
	postgresConfig struct {
		raw rawConfig
	}
)

var _ def.Config = (*Config)(nil)

func NewConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("load .env: %w", err)
	}

	var raw rawConfig
	if err := cenv.Parse(&raw); err != nil {
		return nil, err
	}

	return &Config{raw: raw}, nil
}

func (c *Config) Logger() def.LoggerConfig     { return &loggerConfig{raw: c.raw} }
func (c *Config) HTTP() def.HTTPConfig         { return &httpConfig{raw: c.raw} }
func (c *Config) Auth() def.AuthConfig         { return &authConfig{raw: c.raw} }
func (c *Config) Postgres() def.PostgresConfig { return &postgresConfig{raw: c.raw} }

func (c *loggerConfig) Level() string  { return c.raw.LoggerLevel }
func (c *loggerConfig) AsJSON() bool   { return c.raw.LoggerAsJSON }
func (c *authConfig) Username() string { return c.raw.AuthUsername }
func (c *authConfig) Password() string { return c.raw.AuthPassword }
func (c *httpConfig) Addr() string {
	return net.JoinHostPort(c.raw.HTTPHost, c.raw.HTTPPort)
}

func (c *postgresConfig) DSN() string           { return c.raw.PostgresDSN }
func (c *postgresConfig) MigrationsDir() string { return c.raw.MigrationsDir }
func (c *postgresConfig) AutoMigrate() bool     { return c.raw.AutoMigrate }
