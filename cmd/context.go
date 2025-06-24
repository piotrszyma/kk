package cmd

import (
	"log"

	"github.com/piotrszyma/kk/internal/cli"
	"github.com/piotrszyma/kk/internal/k8s"
	"github.com/spf13/cobra"
)

// contextCmd represents the context command
var contextCmd = &cobra.Command{
	Use:   "context",
	Short: "Change context interactively",
	Long: `The 'context' command provides an interactive way to view and switch your
current Kubernetes context. It enhances the standard 'kubectl config use-context'
by offering a fuzzy finder interface and support for custom aliases, making it
easier to manage numerous or complex context names.

Configuration:
To define custom aliases, create or edit the configuration file at:
'~/.config/kk/config.yaml'
`,
	Aliases: []string{"c", "ctx"},
	Run: func(cmd *cobra.Command, args []string) {
		// TODO(pszyma): Ability to customize kubeConfig path.
		kubeConfigPath := ""

		k8sConfig, err := k8s.LoadConfig(kubeConfigPath)
		if err != nil {
			log.Fatalf("failed to load k8s config: %v", err)
		}

		config, err := cli.LoadConfig()
		if err != nil {
			log.Fatalf("failed to load kk config: %v", err)
		}

		if err := cli.ChangeContext(config, k8sConfig); err != nil {
			log.Fatalf("context change failed: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(contextCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// contextCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// contextCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
