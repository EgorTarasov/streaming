package config

import (
	"github.com/EgorTarasov/streaming/responser/pkg/infrastructure/minios3"
	"github.com/EgorTarasov/streaming/responser/pkg/infrastructure/postgres"
	"github.com/ilyakaznacheev/cleanenv"
)

type KafkaConf struct {
	PredictTopic  string `yaml:"predict-topic" env:"PREDICT_TOPIC" env-default:"prediction"`
	ResponseTopic string `yaml:"response-topic" env:"RESPONSE_TOPIC" env-default:"responses"`
	ResultTopic   string `yaml:"result-topic" env:"RESULT_TOPIC" env-default:"frames-splitted"`
	Broker        string `yaml:"kafka-broker" env:"KAFKA_BROKER" env-default:"127.0.0.1:9091"`
}

type Config struct {
	Kafka    KafkaConf       `yaml:"kafka"`
	Postgres postgres.Config `yaml:"postgres"`
	S3       minios3.Config  `yaml:"s3"`
}

// MustNew создание конфигурации для запуска сервиса
func MustNew(confPath string) Config {
	var cfg Config
	err := cleanenv.ReadConfig(confPath, &cfg)
	if err != nil {
		panic(err)
	}
	return cfg
}
