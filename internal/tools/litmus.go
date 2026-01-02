package tools

import (
	"context"

	eirenyx "github.com/EirenyxK8s/eirenyx/api/v1alpha1"
)

type LitmusService struct {
}

func (l *LitmusService) Name() string {
	return "litmus"
}

func (l *LitmusService) EnsureInstalled(ctx context.Context, tool *eirenyx.Tool) error {
	return nil
}

func (l *LitmusService) EnsureUninstalled(ctx context.Context, tool *eirenyx.Tool) error {
	return nil
}

func (l *LitmusService) CheckHealth(ctx context.Context, tool *eirenyx.Tool) (bool, error) {
	return true, nil
}
