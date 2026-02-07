package litmus

import (
	"context"
	"fmt"

	"github.com/EirenyxK8s/eirenyx/api/litmus"
	eirenyx "github.com/EirenyxK8s/eirenyx/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	managedByLabelKey     = "app.kubernetes.io/managed-by"
	managedByLabelVal     = "eirenyx"
	policyNameLabelKey    = "eirenyx.eirenyx/policy-name"
	policyTypeLabelKey    = "eirenyx.eirenyx/policy-type"
	litmusExperimentLabel = "eirenyx.eirenyx/litmus-experiment"
)

type Engine struct {
	Client client.Client
	Scheme *runtime.Scheme
}

func (e *Engine) Validate(policy *eirenyx.Policy) error {
	if policy.Spec.Type != eirenyx.PolicyTypeLitmus {
		return fmt.Errorf("litmus engine received unsupported policy type: %s", policy.Spec.Type)
	}

	if policy.Spec.Litmus == nil {
		return fmt.Errorf("spec.litmus is required for type=litmus")
	}

	if len(policy.Spec.Litmus.Experiments) == 0 {
		return fmt.Errorf("spec.litmus.experiments must contain at least one experiment")
	}

	for i, exp := range policy.Spec.Litmus.Experiments {
		if exp.Name == "" {
			return fmt.Errorf("litmus.experiments[%d].name is required", i)
		}
		if exp.ExperimentRef == "" {
			return fmt.Errorf("litmus.experiments[%d].experimentRef is required", i)
		}
		if exp.AppInfo.AppNamespace == "" ||
			exp.AppInfo.AppLabel == "" ||
			exp.AppInfo.AppKind == "" {
			return fmt.Errorf("litmus.experiments[%d].appInfo is incomplete", i)
		}
	}

	return nil
}

func (e *Engine) Reconcile(ctx context.Context, policy *eirenyx.Policy) error {
	for _, exp := range policy.Spec.Litmus.Experiments {

		ns := policy.Namespace
		if exp.TargetNamespace != "" {
			ns = exp.TargetNamespace
		}

		engine := &litmus.ChaosEngine{
			ObjectMeta: metav1.ObjectMeta{
				Name:      getChaosEngineName(policy, exp.Name),
				Namespace: ns,
			},
		}

		_, err := controllerutil.CreateOrUpdate(ctx, e.Client, engine, func() error {
			engine.Labels = map[string]string{
				managedByLabelKey:     managedByLabelVal,
				policyNameLabelKey:    policy.Name,
				policyTypeLabelKey:    string(policy.Spec.Type),
				litmusExperimentLabel: exp.Name,
			}

			engine.Spec = litmus.ChaosEngineSpec{
				EngineState: "active",
				AppInfo: litmus.ApplicationParams{
					Appns:    exp.AppInfo.AppNamespace,
					Applabel: exp.AppInfo.AppLabel,
					AppKind:  exp.AppInfo.AppKind,
				},
				Experiments: []litmus.ExperimentList{
					{
						Name: exp.ExperimentRef,
						Spec: litmus.ExperimentAttributes{
							Components: litmus.ExperimentComponents{
								ENV: buildEnvVars(exp),
							},
						},
					},
				},
			}

			return nil
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func buildEnvVars(exp eirenyx.LitmusExperiment) []corev1.EnvVar {
	envs := make([]corev1.EnvVar, 0)

	if exp.Duration != "" {
		envs = append(envs, corev1.EnvVar{
			Name:  "TOTAL_CHAOS_DURATION",
			Value: exp.Duration,
		})
	}

	if exp.Mode != "" {
		envs = append(envs, corev1.EnvVar{
			Name:  "CHAOS_MODE",
			Value: exp.Mode,
		})
	}

	for k, v := range exp.Parameters {
		envs = append(envs, corev1.EnvVar{
			Name:  k,
			Value: v,
		})
	}

	return envs
}

func (e *Engine) Cleanup(ctx context.Context, policy *eirenyx.Policy) error {
	namespaces := map[string]struct{}{
		policy.Namespace: {},
	}

	for _, exp := range policy.Spec.Litmus.Experiments {
		if exp.TargetNamespace != "" {
			namespaces[exp.TargetNamespace] = struct{}{}
		}
	}

	for ns := range namespaces {
		matchedLabels := client.MatchingLabels{
			managedByLabelKey:  managedByLabelVal,
			policyNameLabelKey: policy.Name,
		}

		var engineList litmus.ChaosEngineList

		if err := e.Client.List(ctx, &engineList, client.InNamespace(ns), matchedLabels); err != nil {
			return err
		}

		for i := range engineList.Items {
			if err := e.Client.Delete(ctx, &engineList.Items[i]); client.IgnoreNotFound(err) != nil {
				return err
			}
		}
	}

	return nil
}

func (e *Engine) GenerateReport(ctx context.Context, policy *eirenyx.Policy) (string, error) {
	return fmt.Sprintf("litmus-chaos-report-%s", policy.Name), nil
}

func getChaosEngineName(policy *eirenyx.Policy, experimentName string) string {
	return fmt.Sprintf("eirenyx-litmus-%s-%s", policy.Name, experimentName)
}
