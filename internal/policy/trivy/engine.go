package trivy

import (
	"context"
	"fmt"

	eirenyx "github.com/EirenyxK8s/eirenyx/api/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
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
	if policy.Spec.Type != eirenyx.PolicyTypeTrivy {
		return fmt.Errorf("trivy engine received unsupported policy type: %s", policy.Spec.Type)
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
		existing := &batchv1.Job{}
		err := e.Client.Get(ctx, client.ObjectKey{
			Name:      getScanJobName(policy, scan.Name),
			Namespace: policy.Namespace,
		}, existing)

		// If the Job already exists, no need to create it again
		if err == nil {
			return nil
		}

		if !apierrors.IsNotFound(err) {
			return err
		}

		backoffLimit := int32(0)
		ttl := int32(300)

		job := &batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:      getScanJobName(policy, scan.Name),
				Namespace: policy.Namespace,
				Labels: map[string]string{
					managedByLabelKey:     managedByLabelVal,
					policyNameLabelKey:    policy.Name,
					policyTypeLabelKey:    string(policy.Spec.Type),
					trivyScanNameLabelKey: scan.Name,
				},
			},
			Spec: batchv1.JobSpec{
				BackoffLimit:            &backoffLimit,
				TTLSecondsAfterFinished: &ttl,
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							trivyScanNameLabelKey: scan.Name,
						},
					},
					Spec: corev1.PodSpec{
						RestartPolicy: corev1.RestartPolicyNever,
						Containers: []corev1.Container{
							{
								Name:    "trivy",
								Image:   "aquasec/trivy:latest",
								Command: []string{"trivy", "image", scan.Image},
							},
						},
					},
				},
			},
		}

		if err := controllerutil.SetControllerReference(policy, job, e.Scheme); err != nil {
			return err
		}
		return e.Client.Create(ctx, job)
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
	return fmt.Sprintf("eirenyx-trivy-%s-%s", policy.Name, scanName)
}
