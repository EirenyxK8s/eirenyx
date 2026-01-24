package falco

import (
	"context"
	"fmt"
	"sort"
	"strings"

	eirenyx "github.com/EirenyxK8s/eirenyx/api/v1alpha1"
	"github.com/EirenyxK8s/eirenyx/internal/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	rulesConfigMapKey  = "eirenyx_rules.yaml"
	managedByLabelKey  = "app.kubernetes.io/managed-by"
	managedByLabelVal  = "eirenyx"
	policyNameLabelKey = "eirenyx.eirenyx/policy-name"
	policyTypeLabelKey = "eirenyx.eirenyx/policy-type"
)

// Engine renders Eirenyx Falco policies into Falco-consumable resources.
type Engine struct {
	Client client.Client
	Scheme *runtime.Scheme
}

func (e *Engine) Validate(policy *eirenyx.Policy) error {
	if policy.Spec.Base.Type != eirenyx.PolicyTypeFalco {
		return fmt.Errorf("falco engine received unsupported policy type: %s", policy.Spec.Base.Type)
	}
	if policy.Spec.Falco == nil {
		return fmt.Errorf("spec.falco is required for type=falco")
	}
	if len(policy.Spec.Falco.Rules) == 0 {
		return fmt.Errorf("spec.falco.rules must contain at least one rule")
	}
	for i, r := range policy.Spec.Falco.Rules {
		if strings.TrimSpace(r.Name) == "" {
			return fmt.Errorf("falco rule[%d].name is required", i)
		}
		if strings.TrimSpace(r.Condition) == "" {
			return fmt.Errorf("falco rule[%d].condition is required", i)
		}
		if strings.TrimSpace(r.Output) == "" {
			return fmt.Errorf("falco rule[%d].output is required", i)
		}
		if strings.TrimSpace(r.Priority) == "" {
			return fmt.Errorf("falco rule[%d].priority is required", i)
		}
	}
	return nil
}

func (e *Engine) Reconcile(ctx context.Context, policy *eirenyx.Policy) error {
	rendered := renderFalcoRules(policy)

	configMap := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      getConfigMapName(policy),
			Namespace: policy.Namespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, e.Client, &configMap, func() error {
		if configMap.Labels == nil {
			configMap.Labels = map[string]string{}
		}
		configMap.Labels[managedByLabelKey] = managedByLabelVal
		configMap.Labels[policyNameLabelKey] = policy.Name
		configMap.Labels[policyTypeLabelKey] = string(policy.Spec.Base.Type)

		if configMap.Data == nil {
			configMap.Data = map[string]string{}
		}
		configMap.Data[rulesConfigMapKey] = rendered

		return controllerutil.SetControllerReference(policy, &configMap, e.Scheme)
	})

	return err
}

func (e *Engine) Cleanup(ctx context.Context, policy *eirenyx.Policy) error {
	configMap := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      getConfigMapName(policy),
			Namespace: policy.Namespace,
		},
	}
	return client.IgnoreNotFound(e.Client.Delete(ctx, &configMap))
}

func (e *Engine) GenerateReport(ctx context.Context, policy *eirenyx.Policy) (string, error) {
	return fmt.Sprintf("falco-report-%s", policy.Name), nil
}

func getConfigMapName(policy *eirenyx.Policy) string {
	return fmt.Sprintf("eirenyx-falco-policy-%s", policy.Name)
}

func renderFalcoRules(policy *eirenyx.Policy) string {
	// Falco rule file format is YAML with a "rules:" list.
	var builder strings.Builder
	builder.WriteString("rules:\n")

	rules := append([]eirenyx.FalcoRule(nil), policy.Spec.Falco.Rules...)
	sort.Slice(rules, func(i, j int) bool { return rules[i].Name < rules[j].Name })

	for _, rule := range rules {
		selectors := injectTargetSelectors(rule.Condition, policy.Spec.Base.Target)

		builder.WriteString(fmt.Sprintf("- rule: %q\n", rule.Name))
		builder.WriteString(fmt.Sprintf("  desc: %q\n", fmt.Sprintf("Managed by Eirenyx policy %s", policy.Name)))
		builder.WriteString(fmt.Sprintf("  condition: %s\n", utils.ToYamlBlock(selectors, 4)))
		builder.WriteString(fmt.Sprintf("  output: %s\n", utils.ToYamlBlock(rule.Output, 4)))
		builder.WriteString(fmt.Sprintf("  priority: %s\n", rule.Priority))

		if len(rule.Tags) > 0 {
			builder.WriteString("  tags:\n")
			for _, tag := range rule.Tags {
				builder.WriteString(fmt.Sprintf("    - %q\n", tag))
			}
		}
	}

	return builder.String()
}

func injectTargetSelectors(baseCondition string, target eirenyx.PolicyTarget) string {
	condition := strings.TrimSpace(baseCondition)
	var extra []string

	// Namespace scoping
	if len(target.NamespaceSelector) == 1 {
		extra = append(extra, fmt.Sprintf(`k8s.ns.name = "%s"`, target.NamespaceSelector[0]))
	} else if len(target.NamespaceSelector) > 1 {
		extra = append(extra, fmt.Sprintf(`k8s.ns.name in (%s)`, utils.JoinQuoted(target.NamespaceSelector)))
	}

	// Node scoping
	if len(target.NodeSelector) == 1 {
		extra = append(extra, fmt.Sprintf(`k8s.node.name = "%s"`, target.NodeSelector[0]))
	} else if len(target.NodeSelector) > 1 {
		extra = append(extra, fmt.Sprintf(`k8s.node.name in (%s)`, utils.JoinQuoted(target.NodeSelector)))
	}

	if len(extra) == 0 {
		return condition
	}
	return fmt.Sprintf("(%s) and (%s)", condition, strings.Join(extra, " and "))
}
