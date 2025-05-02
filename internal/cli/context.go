package cli

import (
	"cmp"
	"fmt"
	"os"
	"slices"

	"github.com/piotrszyma/kk/internal/k8s"
	"github.com/piotrszyma/kk/internal/tui"
	"github.com/pkg/errors"
)

type ContextOption struct {
	label       string
	contextName string
	alias       string
	isCurrent   bool
}

func (c ContextOption) Label() string {
	return c.label
}

func (c ContextOption) Preview() string {
	var aliasSuffix string
	if c.alias != "" {
		aliasSuffix = fmt.Sprintf(" (aliased \"%s\")", c.alias)
	}
	return fmt.Sprintf("%s%s", c.contextName, aliasSuffix)
}

func contextAliases(
	config Config,
) map[string][]string {
	ctxNameToAliases := map[string][]string{}
	for _, alias := range config.ContextConfig.Aliases {
		if _, ok := ctxNameToAliases[alias.Name]; !ok {
			ctxNameToAliases[alias.Name] = []string{}
		}

		ctxNameToAliases[alias.Name] = append(ctxNameToAliases[alias.Name], alias.Alias)
	}

	return ctxNameToAliases
}

func contextOptsSorted(
	k8sConfig *k8s.ApiConfig,
	ctxNameToAliases map[string][]string,
) []ContextOption {
	opts := []ContextOption{}

	for contextName := range k8sConfig.Contexts {
		isCurrent := contextName == k8sConfig.CurrentContext
		aliases := ctxNameToAliases[contextName]

		for _, alias := range aliases {
			opts = append(opts, ContextOption{
				label:       alias,
				contextName: contextName,
				isCurrent:   isCurrent,
				alias:       alias,
			})
		}

		opts = append(opts, ContextOption{
			label:       contextName,
			contextName: contextName,
			isCurrent:   isCurrent,
		})
	}

	slices.SortStableFunc(opts, func(o1, o2 ContextOption) int {
		return cmp.Compare(o1.Label(), o2.Label())
	})

	return opts
}

func ChangeContext(
	config Config,
	k8sConfig *k8s.ApiConfig,
) error {
	ctxNameToAliases := contextAliases(config)

	opts := contextOptsSorted(k8sConfig, ctxNameToAliases)

	optSelected, err := tui.OptionPicker(opts)
	if err != nil {
		if tui.IsAbortError(err) {
			return nil
		}

		return errors.Wrapf(err, "failed to select option")
	}

	if err := k8s.SetCurrentContext(k8sConfig, optSelected.contextName); err != nil {
		return errors.Wrapf(err, "failed to set k8s context")
	}

	var aliasSuffix string
	if optSelected.alias != "" {
		aliasSuffix = fmt.Sprintf(" (aliased: \"%s\")", optSelected.alias)
	}

	fmt.Fprintf(os.Stdout, "Switched to context \"%s\"%s.\n", optSelected.contextName, aliasSuffix)

	return nil
}
