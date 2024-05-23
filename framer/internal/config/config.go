package config

import "github.com/ilyakaznacheev/cleanenv"

type KafkaConf struct {
	CmdTopic      string `yaml:"cmd-topic" env:"CMD_TOPIC" env-default:"frames"`
	ResponseTopic string `yaml:"response-topic" env:"RESPONSE_TOPIC" env-default:"responses"`
	ResultTopic   string `yaml:"result-topic" env:"RESULT_TOPIC" env-default:"frames-splitted"`
	Broker        string `yaml:"kafka-broker" env:"KAFKA_BROKER" env-default:"localhost:9091"` // FIXME: move into config.yaml
}

type Config struct {
	Kafka KafkaConf
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
