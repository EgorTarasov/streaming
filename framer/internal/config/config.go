package config

import (
	"github.com/EgorTarasov/streaming/framer/pkg/infrastructure/minios3"
	"github.com/ilyakaznacheev/cleanenv"
)

type KafkaConf struct {
	CmdTopic      string `yaml:"cmd-topic" env:"CMD_TOPIC" env-default:"frames"`
	ResponseTopic string `yaml:"response-topic" env:"RESPONSE_TOPIC" env-default:"responses"`
	ResultTopic   string `yaml:"result-topic" env:"RESULT_TOPIC" env-default:"frames-splitted"`
	Broker        string `yaml:"kafka-broker" env:"KAFKA_BROKER"` // FIXME: move into config.yaml
}

type Config struct {
	Kafka KafkaConf      `yaml:"kafka"`
	S3    minios3.Config `yaml:"s3"`
}

// MustNew создание конфигурации для запуска сервиса
func MustNew(cfgPath string) Config {
	var cfg Config
	err := cleanenv.ReadConfig(cfgPath, &cfg)
	// TODO: solve config nightmare
	cfg.S3.Buckets = []minios3.Bucket{minios3.Bucket{
		Name:   "dev",
		Region: "us-east-1",
		Lock:   false,
	}}
	if err != nil {
		panic(err)
	}
	return cfg
}
