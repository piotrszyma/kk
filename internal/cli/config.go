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

func loadConfigOrCreateEmpty(configPath string) (Config, error) {
	// Create a new Config struct
	var config Config

	// Try to read the configuration file
	yamlFile, err := os.ReadFile(configPath)
	if err != nil {
		// Check if the error is due to the file not existing
		if os.IsNotExist(err) {
			return createEmptyConfigFile(configPath)
		}

		// Return other file reading errors
		return Config{}, fmt.Errorf("failed to read config file '%s': %w", configPath, err)
	}

	// Unmarshal the YAML data into the struct
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal YAML from '%s': %w", configPath, err)
	}

	return config, nil
}

func createEmptyConfigFile(configPath string) (Config, error) {
	// Create the necessary directories if they don't exist
	err := os.MkdirAll(filepath.Dir(configPath), 0755)
	if err != nil {
		return Config{}, fmt.Errorf("failed to create config directory '%s': %w", filepath.Dir(configPath), err)
	}

	// Create a new empty Config struct
	config := Config{}

	// Marshal the empty config to YAML
	yamlData, err := yaml.Marshal(config)
	if err != nil {
		return Config{}, fmt.Errorf("failed to marshal empty config: %w", err)
	}

	// Write the empty config to the file
	err = os.WriteFile(configPath, yamlData, 0644)
	if err != nil {
		return Config{}, fmt.Errorf("failed to write empty config to '%s': %w", configPath, err)
	}

	return config, nil
}

func LoadConfig() (Config, error) {
	dirname, err := os.UserHomeDir()
	if err != nil {
		return Config{}, fmt.Errorf("failed to read user home dir: %w", err)
	}

	configPath := filepath.Join(dirname, ".config", "kk", "config.yaml")

	return loadConfigOrCreateEmpty(configPath)
}
