package trivy

import (
	"context"
	"fmt"

	eirenyx "github.com/EirenyxK8s/eirenyx/api/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
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
	trivyScanNameLabelKey = "eirenyx.eirenyx/trivy-scan-name"
)

type Engine struct {
	Client client.Client
	Scheme *runtime.Scheme
}

func (e *Engine) Validate(policy *eirenyx.Policy) error {
	if policy.Spec.Base.Type != eirenyx.PolicyTypeTrivy {
		return fmt.Errorf("trivy engine received unsupported policy type: %s", policy.Spec.Base.Type)
	}

	if policy.Spec.Trivy == nil {
		return fmt.Errorf("spec.trivy is required for type=trivy")
	}

	if len(policy.Spec.Trivy.Scans) == 0 {
		return fmt.Errorf("spec.trivy.scans must contain at least one scan")
	}

	for i, scan := range policy.Spec.Trivy.Scans {
		if scan.Name == "" {
			return fmt.Errorf("trivy.scans[%d].name is required", i)
		}
		if scan.Image == "" {
			return fmt.Errorf("trivy.scans[%d].image is required", i)
		}
	}

	return nil
}

func (e *Engine) Reconcile(ctx context.Context, policy *eirenyx.Policy) error {
	for _, scan := range policy.Spec.Trivy.Scans {

		job := batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:      getScanJobName(policy, scan.Name),
				Namespace: policy.Namespace,
			},
		}

		_, err := controllerutil.CreateOrUpdate(ctx, e.Client, &job, func() error {
			if job.Labels == nil {
				job.Labels = map[string]string{}
			}

			job.Labels[managedByLabelKey] = managedByLabelVal
			job.Labels[policyNameLabelKey] = policy.Name
			job.Labels[policyTypeLabelKey] = string(policy.Spec.Base.Type)
			job.Labels[trivyScanNameLabelKey] = scan.Name

			backoffLimit := int32(0)
			ttl := int32(300)

			job.Spec = batchv1.JobSpec{
				BackoffLimit:            &backoffLimit,
				TTLSecondsAfterFinished: &ttl,
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: job.Labels,
					},
					Spec: corev1.PodSpec{
						RestartPolicy: corev1.RestartPolicyNever,
						Containers: []corev1.Container{
							{
								Name:            "image",
								Image:           scan.Image,
								ImagePullPolicy: corev1.PullAlways,
								Command:         []string{"sh", "-c", "exit 0"},
							},
						},
					},
				},
			}
			return controllerutil.SetControllerReference(policy, &job, e.Scheme)
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func (e *Engine) Cleanup(ctx context.Context, policy *eirenyx.Policy) error {
	for _, scan := range policy.Spec.Trivy.Scans {
		job := batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:      getScanJobName(policy, scan.Name),
				Namespace: policy.Namespace,
			},
		}

		if err := e.Client.Delete(ctx, &job); client.IgnoreNotFound(err) != nil {
			return err
		}
	}
	return nil
}

func (e *Engine) GenerateReport(ctx context.Context, policy *eirenyx.Policy) (string, error) {
	return "trivy-report-name", nil
}

func getScanJobName(policy *eirenyx.Policy, scanName string) string {
	return fmt.Sprintf(
		"eirenyx-trivy-%s-%s-%d",
		policy.Name,
		scanName,
		policy.Generation,
	)
}
