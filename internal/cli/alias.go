package cli

import (
	"k8s.io/client-go/tools/clientcmd/api"
)

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
