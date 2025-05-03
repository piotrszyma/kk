package cmd

import (
	"fmt"

	"github.com/piotrszyma/kk/internal/cli"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// contextResolveCmd represents the resolve command
var contextResolveCmd = &cobra.Command{
	Use:   "resolve",
	Short: "Resolve context by name alias",
	Long: `Accepts single positional argument - context name alias and prints
real context name.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		aliasToResolve := args[0]

		config, err := cli.LoadConfig()
		if err != nil {
			return errors.Wrapf(err, "failed to load kk config")
		}

		for _, alias := range config.ContextConfig.Aliases {
			if alias.Alias == aliasToResolve {
				fmt.Print(alias.Name)
				return nil
			}
		}

		return errors.Errorf("failed to resolve alias %s", args)
	},
}

func init() {
	contextCmd.AddCommand(contextResolveCmd)
}
