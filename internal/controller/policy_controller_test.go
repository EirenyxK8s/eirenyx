package controller

import (
	"context"
	"sync"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	eirenyxlitmus "github.com/EirenyxK8s/eirenyx/api/litmus"
	eirenyxv1alpha1 "github.com/EirenyxK8s/eirenyx/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Policy Controller", func() {
	Context("Policy reconciliation lifecycle", func() {
		const (
			policyName = "test-policy"
			namespace  = "default"
		)

		ctx := context.Background()

		key := types.NamespacedName{
			Name:      policyName,
			Namespace: namespace,
		}

		var addSchemesOnce sync.Once

		BeforeEach(func() {
			By("registering required schemes")
			addSchemesOnce.Do(func() {
				Expect(eirenyxlitmus.AddToScheme(k8sClient.Scheme())).To(Succeed())
			})

			By("creating Tool dependency")
			tool := &eirenyxv1alpha1.Tool{
				ObjectMeta: metav1.ObjectMeta{
					Name:      string(eirenyxv1alpha1.PolicyTypeLitmus),
					Namespace: namespace,
				},
			}
			Expect(k8sClient.Create(ctx, tool)).To(Succeed())

			By("creating Policy resource")
			policy := &eirenyxv1alpha1.Policy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      policyName,
					Namespace: namespace,
				},
				Spec: eirenyxv1alpha1.PolicySpec{
					Type:    eirenyxv1alpha1.PolicyTypeLitmus,
					Enabled: true,
					Litmus: &eirenyxv1alpha1.LitmusPolicySpec{
						Experiments: []eirenyxv1alpha1.LitmusExperiment{
							{
								Name:          "cpu-hog",
								ExperimentRef: "pod-cpu-hog",
								AppInfo: eirenyxv1alpha1.LitmusAppInfo{
									AppNamespace: "default",
									AppLabel:     "app=test",
									AppKind:      "deployment",
								},
							},
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, policy)).To(Succeed())
		})

		AfterEach(func() {
			policy := &eirenyxv1alpha1.Policy{}
			if err := k8sClient.Get(ctx, key, policy); err == nil {
				_ = k8sClient.Delete(ctx, policy)
			}

			tool := &eirenyxv1alpha1.Tool{}
			if err := k8sClient.Get(ctx, types.NamespacedName{
				Name:      string(eirenyxv1alpha1.PolicyTypeLitmus),
				Namespace: namespace,
			}, tool); err == nil {
				_ = k8sClient.Delete(ctx, tool)
			}
		})

	})
})
