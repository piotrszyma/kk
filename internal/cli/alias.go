package cli

import (
	"github.com/piotrszyma/kk/internal/k8s"
)

// ContextWithAlias may have alias.
type ContextWithAlias struct {
	Name      string
	Context   *k8s.ApiContext
	IsCurrent bool

	// Alias, empty indicates no alias.
	Alias string
}

func (c ContextWithAlias) HasAlias() bool {
	return c.Alias != ""
}
