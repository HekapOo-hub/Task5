package config

import (
	"fmt"
	"github.com/caarlos0/env/v6"
)

type RabbitConfig struct {
	Host string `env:"rabbitHost" envDefault:"localhost"`
	Port string `env:"rabbitPort" envDefault:"5672"`
}

func (c *RabbitConfig) GetURL() string {
	//amqp://guest:guest@localhost:5672/
	return fmt.Sprintf("amqp://guest:guest@%s:%s/", c.Host, c.Port)
}

func GetRabbitConfig() (*RabbitConfig, error) {
	cfg := RabbitConfig{}
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("error with parsing env variables in rabbit config %w", err)
	}
	return &cfg, nil
}
