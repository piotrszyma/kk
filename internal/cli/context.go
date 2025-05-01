package cli

import (
	"errors"
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/ktr0731/go-fuzzyfinder"
	"k8s.io/client-go/tools/clientcmd"
)

func ChangeContext() {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	// Alternatively, specify a path directly:
	// loadingRules.ExplicitPath = "/path/to/your/.kube/config"

	config, err := LoadConfig()
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
