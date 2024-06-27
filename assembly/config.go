package assembly

import "time"

type Config struct {
	PostresDsn string `envconfig:"POSTGRES_DSN"`
	HttpPort   string `envconfig:"HTTP_PORT"`

	DbMaxOpenConns    int           `envconfig:"DB_MAX_OPEN_CONNS" default:"4"`
	DbMaxIdleConns    int           `envconfig:"DB_MAX_IDLE_CONNS" default:"4"`
	DbConnMaxLifetime time.Duration `envconfig:"DB_CONN_MAX_LIFETIME" default:"5m"`
	DbConnMaxIdleTime time.Duration `envconfig:"DB_CONN_MAX_IDLE_TIME" default:"5m"`
}

func NewConfig() (Config, error) {
	conf := Config{}

	return conf, nil
}
