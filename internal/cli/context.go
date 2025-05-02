package cli

import (
	"cmp"
	"fmt"
	"os"
	"slices"

	"github.com/piotrszyma/kk/internal/tui"
	"github.com/pkg/errors"
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

type ContextOption struct {
	label       string
	contextName string
	alias       string
	context     *api.Context
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

// TODO(pszyma): Make this function accept `Config`.
func ChangeContext(
	kubeConfigPath string,
) error {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if kubeConfigPath != "" {
		loadingRules.ExplicitPath = kubeConfigPath
	}

	config, err := LoadConfig()
	if err != nil {
		return errors.Wrapf(err, "failed to load kk config")
	}

	k8sConfig, err := loadingRules.Load()
	if err != nil {
		return errors.Wrapf(err, "failed to load k8s config")
	}

	ctxToAlias := map[string]string{}
	for _, alias := range config.ContextConfig.Aliases {
		ctxToAlias[alias.Name] = alias.Alias
	}

	opts := []ContextOption{}

	for contextName, context := range k8sConfig.Contexts {
		isCurrent := contextName == k8sConfig.CurrentContext
		alias := ctxToAlias[contextName]

		if alias != "" {
			opts = append(opts, ContextOption{
				label:       alias,
				contextName: contextName,
				context:     context,
				isCurrent:   isCurrent,
				alias:       alias,
			})
		}

		opts = append(opts, ContextOption{
			label:       contextName,
			contextName: contextName,
			context:     context,
			isCurrent:   isCurrent,
			alias:       alias,
		})
	}

	slices.SortStableFunc(opts, func(o1, o2 ContextOption) int {
		return cmp.Compare(o1.Label(), o2.Label())
	})

	optSelected, err := tui.OptionPicker(opts)
	if err != nil {
		return errors.Wrapf(err, "failed to select option")
	}

	err = setK8sContext(k8sConfig, optSelected.contextName)
	if err != nil {
		return errors.Wrapf(err, "failed to set k8s context")
	}

	var aliasSuffix string
	if optSelected.alias != "" {
		aliasSuffix = fmt.Sprintf(" (aliased: \"%s\")", optSelected.alias)
	}

	fmt.Fprintf(os.Stdout, "Switched to context \"%s\"%s.\n", optSelected.contextName, aliasSuffix)

	return nil
}
