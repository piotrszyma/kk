package k8s

import (
	"github.com/pkg/errors"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

func LoadConfig(kubeConfigPath string) (*api.Config, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if kubeConfigPath != "" {
		loadingRules.ExplicitPath = kubeConfigPath
	}

	k8sConfig, err := loadingRules.Load()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load k8s config")
	}

	return k8sConfig, nil
}

func SetCurrentContext(config *api.Config, contextName string) error {
	configAccess := clientcmd.NewDefaultPathOptions()
	config.CurrentContext = contextName
	err := clientcmd.ModifyConfig(configAccess, *config, true)
	if err != nil {
		return err
	}

	return nil
}
