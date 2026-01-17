package tools

import (
	"context"

	eirenyx "github.com/EirenyxK8s/eirenyx/api/v1alpha1"
)

type ToolService interface {
	Name() string

	EnsureInstalled(ctx context.Context, tool *eirenyx.Tool) error

	EnsureUninstalled(ctx context.Context, tool *eirenyx.Tool) error

	CheckHealth(ctx context.Context, tool *eirenyx.Tool) bool
}
