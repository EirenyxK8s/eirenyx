package falco

import (
	"context"
	"fmt"

	eirenyx "github.com/EirenyxK8s/eirenyx/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// Engine processes Falco-based Eirenyx policies.
type Engine struct {
	Client client.Client
	Scheme *runtime.Scheme
}

func (e *Engine) Validate(policy *eirenyx.Policy) error {
	if policy.Spec.Type != eirenyx.PolicyTypeFalco {
		return fmt.Errorf("falco engine received unsupported policy type: %s", policy.Spec.Type)
	}
	if policy.Spec.Falco == nil {
		return fmt.Errorf("falco policy spec must be defined")
	}

	observe := policy.Spec.Falco.Observe
	if observe.RuleRef == nil && observe.RuleSelector == nil {
		return fmt.Errorf("falco policy must define either ruleRef or ruleSelector")
	}
	if observe.RuleRef != nil && observe.RuleSelector != nil {
		return fmt.Errorf("falco policy cannot define both ruleRef and ruleSelector")
	}
	if observe.RuleRef != nil && observe.RuleRef.Name == "" {
		return fmt.Errorf("ruleRef.name must not be empty")
	}
	return nil
}

func (e *Engine) Reconcile(ctx context.Context, policy *eirenyx.Policy) error {
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      falcoPolicyConfigMapName(policy),
			Namespace: policy.Namespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, e.Client, cm, func() error {
		if err := controllerutil.SetControllerReference(policy, cm, e.Scheme); err != nil {
			return err
		}

		specBytes, err := json.Marshal(policy.Spec.Falco)
		if err != nil {
			return err
		}

		if cm.Labels == nil {
			cm.Labels = map[string]string{}
		}
		cm.Labels["eirenyx.io/type"] = "falco-policy"
		cm.Labels["eirenyx.io/policy"] = policy.Name

		if cm.Data == nil {
			cm.Data = map[string]string{}
		}
		cm.Data["falcoPolicy.json"] = string(specBytes)

		return nil
	})

	return err
}

func (e *Engine) Cleanup(ctx context.Context, policy *eirenyx.Policy) error {
	cm := &corev1.ConfigMap{}
	cmKey := types.NamespacedName{
		Name:      falcoPolicyConfigMapName(policy),
		Namespace: policy.Namespace,
	}

	if err := e.Client.Get(ctx, cmKey, cm); err == nil {
		if err := e.Client.Delete(ctx, cm); err != nil {
			return err
		}
	}

	reportList := &eirenyx.PolicyReportList{}
	if err := e.Client.List(ctx, reportList, client.InNamespace(policy.Namespace)); err != nil {
		return err
	}

	for i := range reportList.Items {
		report := &reportList.Items[i]

		if report.Spec.PolicyRef.Name == policy.Name {
			if err := e.Client.Delete(ctx, report); err != nil {
				if apierrors.IsNotFound(err) {
					return nil
				}
				return err
			}
		}
	}

	return nil
}

func (e *Engine) GenerateReport(ctx context.Context, policy *eirenyx.Policy) (*eirenyx.PolicyReport, error) {
	// Generate report content, here using policy.Name for naming purposes
	reportName := fmt.Sprintf("falco-report-%s", policy.Name)

	// Create the PolicyReport object
	report := &eirenyx.PolicyReport{
		ObjectMeta: metav1.ObjectMeta{
			Name:      reportName,
			Namespace: policy.Namespace,
		},
		Spec: eirenyx.PolicyReportSpec{
			PolicyRef: eirenyx.PolicyReference{
				Name:       policy.Name,
				Generation: policy.Generation,
			},
			Type: policy.Spec.Type,
		},
		Status: eirenyx.PolicyReportStatus{
			Phase: eirenyx.ReportPending,
		},
	}

	// Optionally, add details or other information to the report (e.g., Summary, etc.)
	// report.Status.Summary = ...

	return report, nil
}

func falcoPolicyConfigMapName(policy *eirenyx.Policy) string {
	return fmt.Sprintf("eirenyx-falco-policy-%s", policy.Name)
}
