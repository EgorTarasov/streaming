package config

import (
	"github.com/EgorTarasov/streaming/responser/pkg/infrastructure/postgres"
	"github.com/ilyakaznacheev/cleanenv"
)

type KafkaConf struct {
	PredictTopic  string `yaml:"predict-topic" env:"PREDICT_TOPIC" env-default:"frames"`
	ResponseTopic string `yaml:"response-topic" env:"RESPONSE_TOPIC" env-default:"responses"`
	ResultTopic   string `yaml:"result-topic" env:"RESULT_TOPIC" env-default:"frames-splitted"`
	Broker        string `yaml:"kafka-broker" env:"KAFKA_BROKER" env-default:"kafka1:29091"`
}

type Config struct {
	Kafka    KafkaConf       `yaml:"kafka"`
	Postgres postgres.Config `yaml:"postgres"`
}

// MustNew создание конфигурации для запуска сервиса
func MustNew() Config {
	var cfg Config
	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		panic(err)
	}
	return cfg
}
