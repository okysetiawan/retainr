package config

import "time"

type NatsConfig struct {
	Servers  []string      `mapstructure:"servers"`
	User     string        `mapstructure:"user"`
	Password string        `mapstructure:"password"`
	Timeout  time.Duration `mapstructure:"timeout"`
}
