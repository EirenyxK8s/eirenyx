package litmus

import (
	"context"
	"fmt"

	eirenyx "github.com/EirenyxK8s/eirenyx/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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
	if policy.Spec.Base.Type != eirenyx.PolicyTypeLitmus {
		return fmt.Errorf("litmus engine received unsupported policy type: %s", policy.Spec.Base.Type)
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
		targetNamespace := policy.Namespace
		if exp.TargetNamespace != "" {
			targetNamespace = exp.TargetNamespace
		}

		engine := &ChaosEngine{
			TypeMeta: ChaosEngineTypeMeta,
			ObjectMeta: metav1.ObjectMeta{
				Name:      getChaosEngineName(policy, exp.Name),
				Namespace: targetNamespace,
				Labels: map[string]string{
					managedByLabelKey:     managedByLabelVal,
					policyNameLabelKey:    policy.Name,
					policyTypeLabelKey:    string(policy.Spec.Base.Type),
					litmusExperimentLabel: exp.Name,
				},
			},
			Spec: ChaosEngineSpec{
				EngineState: "active",
				AppInfo: ChaosAppInfo{
					AppNS:    exp.AppInfo.AppNamespace,
					AppLabel: exp.AppInfo.AppLabel,
					AppKind:  exp.AppInfo.AppKind,
				},
				Experiments: []ChaosExperiment{
					{
						Name: exp.ExperimentRef,
						Spec: ChaosExperimentSpec{
							Components: ChaosComponents{
								Env: buildEnvVars(exp),
							},
						},
					},
				},
			},
		}

		unstructuredChaosEngine, err := engine.ToUnstructured()
		if err != nil {
			return err
		}

		_ = controllerutil.SetControllerReference(policy, unstructuredChaosEngine, e.Scheme)

		if err := e.Client.Patch(ctx, unstructuredChaosEngine, client.Apply, client.ForceOwnership); err != nil {
			return err
		}
	}

	return nil
}

func buildEnvVars(exp eirenyx.LitmusExperiment) []ChaosEnvVar {
	var envs []ChaosEnvVar

	if exp.Duration != "" {
		envs = append(envs, ChaosEnvVar{
			Name:  "TOTAL_CHAOS_DURATION",
			Value: exp.Duration,
		})
	}

	if exp.Mode != "" {
		envs = append(envs, ChaosEnvVar{
			Name:  "CHAOS_MODE",
			Value: exp.Mode,
		})
	}

	for k, v := range exp.Parameters {
		envs = append(envs, ChaosEnvVar{
			Name:  k,
			Value: v,
		})
	}

	return envs
}

func (e *Engine) Cleanup(ctx context.Context, policy *eirenyx.Policy) error {
	for _, exp := range policy.Spec.Litmus.Experiments {

		engine := &unstructured.Unstructured{}
		engine.SetGroupVersionKind(ChaosEngineGVK)
		engine.SetName(getChaosEngineName(policy, exp.Name))
		engine.SetNamespace(policy.Namespace)

		if err := e.Client.Delete(ctx, engine); client.IgnoreNotFound(err) != nil {
			return err
		}
	}

	return nil
}

func (e *Engine) GenerateReport(ctx context.Context, policy *eirenyx.Policy) (string, error) {
	return fmt.Sprintf("litmus-chaos-report-%s", policy.Name), nil
}

func getChaosEngineName(policy *eirenyx.Policy, experimentName string) string {
	return fmt.Sprintf(
		"eirenyx-litmus-%s-%s-%d",
		policy.Name,
		experimentName,
		policy.Generation,
	)
}
