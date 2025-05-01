package cli

import (
	"cmp"
	"errors"
	"fmt"
	"log"
	"os"
	"slices"

	"github.com/ktr0731/go-fuzzyfinder"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

func setK8sContext(k8sConfig *api.Config, ctxName string) error {
	configAccess := clientcmd.NewDefaultPathOptions()
	k8sConfig.CurrentContext = ctxName
	err := clientcmd.ModifyConfig(configAccess, *k8sConfig, true)
	if err != nil {
		return err
	}

	return nil
}

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

	slices.SortStableFunc(ctxWithAlias, func(i, j ContextWithAlias) int {
		iHasAlias := i.HasAlias()
		jHasAlias := j.HasAlias()

		// Primary Sort: Alias (Presence and Value)
		if iHasAlias && jHasAlias {
			// Both have aliases: Compare them
			aliasCmp := cmp.Compare(i.Alias, j.Alias)
			if aliasCmp != 0 {
				return aliasCmp // Sort by alias if different
			}
			// Aliases are the same, fall through to compare by Name
		} else if iHasAlias { // Only i has alias
			return -1 // i comes before j
		} else if jHasAlias { // Only j has alias
			return 1 // j comes before i (so i comes after j)
		}
		// If we reach here, either:
		// 1. Aliases were the same (fell through from the first `if`)
		// 2. Neither had an alias (skipped the `else if` blocks)
		// In both cases, proceed to Secondary Sort by Name.

		// Secondary Sort: Name
		return cmp.Compare(i.Name, j.Name)
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

	ctxSelected := ctxWithAlias[idx]

	setK8sContext(k8sConfig, ctxSelected.Name)

	var aliasSuffix string
	if ctxSelected.HasAlias() {
		aliasSuffix = fmt.Sprintf(" (aliased: \"%s\")", ctxSelected.Alias)
	}

	fmt.Fprintf(os.Stdout, "Switched to context \"%s\"%s.\n", ctxSelected.Name, aliasSuffix)
}
