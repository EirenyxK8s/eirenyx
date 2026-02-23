package controller

import (
	"fmt"

	eirenyx "github.com/EirenyxK8s/eirenyx/api/v1alpha1"
	"github.com/EirenyxK8s/eirenyx/internal/client/k8s"
	"github.com/EirenyxK8s/eirenyx/internal/policy"
	"github.com/EirenyxK8s/eirenyx/internal/policy/falco"
	"github.com/EirenyxK8s/eirenyx/internal/policy/litmus"
	"github.com/EirenyxK8s/eirenyx/internal/policy/trivy"
	"github.com/EirenyxK8s/eirenyx/internal/report"
	"github.com/EirenyxK8s/eirenyx/internal/tools"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Dependencies struct {
	Client    client.Client
	Scheme    *runtime.Scheme
	K8sClient *k8s.Client
}

func NewToolService(tool *eirenyx.Tool, deps Dependencies) (tools.ToolService, error) {
	switch tool.Spec.Type {
	case eirenyx.ToolFalco:
		return &tools.FalcoService{
			K8sClient: deps.K8sClient,
		}, nil
	case eirenyx.ToolLitmus:
		return &tools.LitmusService{
			K8sClient: deps.K8sClient,
		}, nil
	case eirenyx.ToolTrivy:
		return &tools.TrivyService{
			K8sClient: deps.K8sClient,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported tool type: %s", tool.Spec.Type)
	}
}

func NewPolicyEngine(p *eirenyx.Policy, deps Dependencies) (policy.Engine, error) {
	switch p.Spec.Type {
	case eirenyx.PolicyTypeFalco:
		return &falco.Engine{
			Client: deps.Client,
			Scheme: deps.Scheme,
		}, nil
	case eirenyx.PolicyTypeTrivy:
		return &trivy.Engine{
			Client: deps.Client,
			Scheme: deps.Scheme,
		}, nil
	case eirenyx.PolicyTypeLitmus:
		return &litmus.Engine{
			Client: deps.Client,
			Scheme: deps.Scheme,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported policy type: %s", p.Spec.Type)
	}
}

func NewReportEngine(r *eirenyx.PolicyReport, deps Dependencies) (report.Handler, error) {
	switch r.Spec.Type {
	case eirenyx.PolicyTypeFalco:
		return &report.FalcoReportHandler{
			Client: deps.Client,
			Scheme: deps.Scheme,
		}, nil
	case eirenyx.PolicyTypeLitmus:
		return &report.LitmusReportHandler{
			Client: deps.Client,
			Scheme: deps.Scheme,
		}, nil
	case eirenyx.PolicyTypeTrivy:
		return &report.TrivyReportHandler{
			Client: deps.Client,
			Scheme: deps.Scheme,
		}, nil
	default:
		return nil, fmt.Errorf("unknown policy type: %s", r.Spec.Type)
	}
}
