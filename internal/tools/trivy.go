package tools

import (
	"context"

	eirenyx "github.com/EirenyxK8s/eirenyx/api/v1alpha1"
)

type TrivyService struct {
}

func (t *TrivyService) Name() string {
	return "trivy"
}

func (t *TrivyService) EnsureInstalled(ctx context.Context, tool *eirenyx.Tool) error {
	return nil
}

func (t *TrivyService) EnsureUninstalled(ctx context.Context, tool *eirenyx.Tool) error {
	return nil
}

func (t *TrivyService) CheckHealth(ctx context.Context, tool *eirenyx.Tool) (bool, error) {
	return true, nil
}
