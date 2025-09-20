package config

type JobConfig struct {
	Name  string `mapstructure:"name"`
	Topic string `mapstructure:"topic"`
}
