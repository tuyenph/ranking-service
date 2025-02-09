package config

import (
	"context"
	"log/slog"
	"os"

	"github.com/sethvargo/go-envconfig"
)

type RedisConfig struct {
	Host     string `env:"HOST, default=localhost:6379"`
	DB       int    `env:"DB, default=0"`
	Username string `env:"USER"`
	Password string `env:"PASSWORD"`
}

type PostgresConfig struct {
	Host     string `env:"HOST, default=localhost:5432"`
	DB       string `env:"DB, default=video_ranking"`
	Username string `env:"USER, default=username"`
	Password string `env:"PASSWORD, default=password"`
}

type ServerConfig struct {
	Port       string         `env:"PORT, default=8080"`
	ListenAddr string         `env:"LISTEN_ADDR, default=0.0.0.0"`
	Redis      RedisConfig    `env:", prefix=REDIS_"`
	Postgres   PostgresConfig `env:", prefix=POSTGRES_"`
}

func MustLoadServerConfigFromEnv() ServerConfig {
	var c ServerConfig
	if err := envconfig.Process(context.Background(), &c); err != nil {
		slog.Error("Failed to process env var", "err", err)
		os.Exit(1)
	}
	return c
}
