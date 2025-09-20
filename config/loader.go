package config

import (
	"fmt"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// LoadConfig loads configuration from file and environment variables
func LoadConfig[T any](configPath string) (T, error) {
	// Optionally load .env if exists
	_ = godotenv.Load()
	v := viper.New()

	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	var config T

	// If a config path is given, load that
	if configPath != "" {
		v.SetConfigFile(configPath)
		// Try reading
		if err := v.ReadInConfig(); err != nil {
			return config, fmt.Errorf("error reading config file: %w", err)
		}
	}

	if err := v.Unmarshal(&config); err != nil {
		return config, fmt.Errorf("unable to decode into config struct: %w", err)
	}

	return config, nil
}
