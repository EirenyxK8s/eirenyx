package controller

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	eirenyxv1alpha1 "github.com/EirenyxK8s/eirenyx/api/v1alpha1"
)

var _ = Describe("Tool Controller", func() {
	Context("When reconciling a Tool resource", func() {
		const (
			toolName  = string(eirenyxv1alpha1.ToolTrivy)
			namespace = "default"
		)

		ctx := context.Background()

		key := types.NamespacedName{
			Name:      toolName,
			Namespace: namespace,
		}

		BeforeEach(func() {
			By("creating the Tool resource")
			tool := &eirenyxv1alpha1.Tool{
				ObjectMeta: metav1.ObjectMeta{
					Name:      toolName, // IMPORTANT: must match spec.type
					Namespace: namespace,
				},
				Spec: eirenyxv1alpha1.ToolSpec{
					Type:      eirenyxv1alpha1.ToolTrivy,
					Enabled:   false, // safe path: no installation logic
					Namespace: "trivy-system",
				},
			}
			Expect(k8sClient.Create(ctx, tool)).To(Succeed())
		})

		AfterEach(func() {
			tool := &eirenyxv1alpha1.Tool{}
			if err := k8sClient.Get(ctx, key, tool); err == nil {
				_ = k8sClient.Delete(ctx, tool)
			}
		})
	})
})
