package assembly

import (
	"log/slog"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	PostresDsn    string `envconfig:"POSTGRES_DSN"`
	HttpPort      string `envconfig:"HTTP_PORT"`
	MigrationsDir string `envconfig:"MIGRATIONS_DIR"`

	DbMaxOpenConns    int           `envconfig:"DB_MAX_OPEN_CONNS" default:"4"`
	DbMaxIdleConns    int           `envconfig:"DB_MAX_IDLE_CONNS" default:"4"`
	DbConnMaxLifetime time.Duration `envconfig:"DB_CONN_MAX_LIFETIME" default:"5m"`
	DbConnMaxIdleTime time.Duration `envconfig:"DB_CONN_MAX_IDLE_TIME" default:"5m"`

	LogLevel slog.Level `envconfig:"LOG_LEVEL" default:"INFO"`
}

func NewConfig() (Config, error) {
	conf := Config{}

	if err := envconfig.Process("", &conf); err != nil {
		return conf, err
	}

	return conf, nil
}
