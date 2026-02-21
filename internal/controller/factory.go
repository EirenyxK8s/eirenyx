package controller

import (
	"fmt"

	eirenyx "github.com/EirenyxK8s/eirenyx/api/v1alpha1"
	"github.com/EirenyxK8s/eirenyx/internal/policy"
	"github.com/EirenyxK8s/eirenyx/internal/policy/falco"
	"github.com/EirenyxK8s/eirenyx/internal/policy/litmus"
	"github.com/EirenyxK8s/eirenyx/internal/policy/trivy"
	"github.com/EirenyxK8s/eirenyx/internal/report"
)

func NewPolicyEngine(policy *eirenyx.Policy, deps Dependencies) (policy.Engine, error) {
	switch policy.Spec.Type {
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
		return nil, fmt.Errorf("unsupported policy type: %s", policy.Spec.Type)
	}
}

func NewReportEngine(policyReport *eirenyx.PolicyReport, deps Dependencies) (report.Handler, error) {
	switch policyReport.Spec.Type {
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
		return nil, fmt.Errorf("unknown policy type: %s", policyReport.Spec.Type)
	}
}
