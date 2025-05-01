package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Alias struct {
	Name  string `yaml:"name"`
	Alias string `yaml:"alias"`
}

type ContextConfig struct {
	Aliases []Alias `yaml:"aliases"`
}

type Config struct {
	ContextConfig ContextConfig `yaml:"context"`
}

func LoadConfig() (Config, error) {
	dirname, err := os.UserHomeDir()
	if err != nil {
		return Config{}, fmt.Errorf("failed to read user home dir: %w", err)
	}

	configPath := filepath.Join(dirname, ".config", "kk", "config.yaml")

	// Create a new Config struct
	config := Config{}

	// Read the configuration file
	yamlFile, err := os.ReadFile(configPath)
	if err != nil {
		// Wrap the error with more context
		return Config{}, fmt.Errorf("failed to read config file '%s': %w", configPath, err)
	}

	// Unmarshal the YAML data into the struct
	// The '&' is important because Unmarshal needs a pointer to fill the struct
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		// Wrap the error with more context
		return Config{}, fmt.Errorf("failed to unmarshal YAML from '%s': %w", configPath, err)
	}

	return config, nil
}
