package tools

import (
	"context"

	eirenyx "github.com/EirenyxK8s/eirenyx/api/v1alpha1"
)

type FalcoService struct {
}

func (f *FalcoService) Name() string {
	return "falco"
}

func (f *FalcoService) EnsureInstalled(ctx context.Context, tool *eirenyx.Tool) error {
	return nil
}

func (f *FalcoService) EnsureUninstalled(ctx context.Context, tool *eirenyx.Tool) error {
	return nil
}

func (f *FalcoService) CheckHealth(ctx context.Context, tool *eirenyx.Tool) (bool, error) {
	return true, nil
}
