package tools

import (
	"context"

	eirenyx "github.com/EirenyxK8s/eirenyx/api/v1alpha1"
)

type ToolService interface {
	Name() string

	EnsureInstalled(ctx context.Context, tool *eirenyx.Tool) error

	EnsureUninstalled(ctx context.Context, tool *eirenyx.Tool) error

	CheckHealth(ctx context.Context, tool *eirenyx.Tool) (bool, error)
}

// resolvedNamespace returns the override if set, otherwise the default.
func resolvedNamespace(override, defaultNS string) string {
	if override != "" {
		return override
	}
	return defaultNS
}
