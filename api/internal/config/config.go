package config

import (
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Port    int          `yaml:"port" env:"PORT"`
	Service Orchestrator `yaml:"orchestrator"`
}

type Orchestrator struct {
	Host string `yaml:"host" env:"ORCHESTRATOR_HOST"`
	Port int    `yaml:"port" env:"ORCHESTRATOR_PORT"`
}

func MustNew() Config {
	var cfg Config
	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		panic(err)
	}
	return cfg
}
