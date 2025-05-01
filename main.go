package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/ktr0731/go-fuzzyfinder"
	"gopkg.in/yaml.v3"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
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

// ContextWithAlias may have alias.
type ContextWithAlias struct {
	Name      string
	Context   *api.Context
	IsCurrent bool

	// Alias, empty indicates no alias.
	Alias string
}

func (c ContextWithAlias) HasAlias() bool {
	return c.Alias != ""
}

func loadConfig() (Config, error) {
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

func main() {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	// Alternatively, specify a path directly:
	// loadingRules.ExplicitPath = "/path/to/your/.kube/config"

	config, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading kk config: %v\n", err)
		os.Exit(1)
	}

	k8sConfig, err := loadingRules.Load() // Directly load the config using rules
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading kubeconfig: %v\n", err)
		os.Exit(1)
	}

	ctxToAlias := map[string]string{}
	for _, alias := range config.ContextConfig.Aliases {
		ctxToAlias[alias.Name] = alias.Alias
	}

	ctxWithAlias := []ContextWithAlias{}
	for contextName, context := range k8sConfig.Contexts {
		isCurrent := contextName == k8sConfig.CurrentContext
		alias := ctxToAlias[contextName]

		ctxWithAlias = append(ctxWithAlias, ContextWithAlias{
			IsCurrent: isCurrent,
			Alias:     alias,
			Name:      contextName,
			Context:   context,
		})
	}

	// Sort ctxWithAlias by Name
	sort.Slice(ctxWithAlias, func(i, j int) bool {
		return ctxWithAlias[i].Name < ctxWithAlias[j].Name
	})

	idx, err := fuzzyfinder.Find(
		ctxWithAlias,
		func(i int) string {
			item := ctxWithAlias[i]

			if item.HasAlias() {
				return fmt.Sprintf("%s (%s)", item.Alias, item.Name)
			} else {
				return item.Name
			}
		},
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i == -1 {
				return ""
			}
			return fmt.Sprintf("Context: %s (%s)",
				ctxWithAlias[i].Name,
				ctxWithAlias[i].Alias,
			)
		}))

	if err != nil {
		if errors.Is(err, fuzzyfinder.ErrAbort) {
			return
		}

		log.Fatal(err)
	}

	ctx := ctxWithAlias[idx]
	configAccess := clientcmd.NewDefaultPathOptions()
	k8sConfig.CurrentContext = ctx.Name
	err = clientcmd.ModifyConfig(configAccess, *k8sConfig, true)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprintf(os.Stdout, "Switched to context \"%s\".\n", ctx.Name)
}
