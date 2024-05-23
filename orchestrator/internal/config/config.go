package config

import (
	"github.com/EgorTarasov/streaming/orchestrator/pkg/db"
	"github.com/EgorTarasov/streaming/orchestrator/pkg/kafka"
	"github.com/EgorTarasov/streaming/orchestrator/pkg/minios3"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Postgres db.Config      `yaml:"postgres"`
	S3       minios3.Config `yaml:"s3"`
	Consumer kafka.Config   `yaml:"kafka"`
}

// MustNew создание конфигурации для запуска сервиса
func MustNew() Config {
	var cfg Config
	err := cleanenv.ReadConfig("config.yaml", &cfg)
	if err != nil {
		panic(err)
	}
	return cfg
}
