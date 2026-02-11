package falco

import (
	"context"
	"fmt"

	eirenyx "github.com/EirenyxK8s/eirenyx/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
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
	key := types.NamespacedName{
		Name:      falcoPolicyConfigMapName(policy),
		Namespace: policy.Namespace,
	}

	if err := e.Client.Get(ctx, key, cm); err != nil {
		return client.IgnoreNotFound(err)
	}

	return e.Client.Delete(ctx, cm)
}

func (e *Engine) GenerateReport(ctx context.Context, policy *eirenyx.Policy) (string, error) {
	return fmt.Sprintf("falco-report-%s", policy.Name), nil
}

func falcoPolicyConfigMapName(policy *eirenyx.Policy) string {
	return fmt.Sprintf("eirenyx-falco-policy-%s", policy.Name)
}
