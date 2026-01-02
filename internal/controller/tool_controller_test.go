package controller

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	eirenyxv1alpha1 "github.com/EirenyxK8s/eirenyx/api/v1alpha1"
)

var _ = Describe("Tool Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}
		tool := &eirenyxv1alpha1.Tool{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind Tool")
			err := k8sClient.Get(ctx, typeNamespacedName, tool)
			if err != nil && errors.IsNotFound(err) {
				resource := &eirenyxv1alpha1.Tool{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: eirenyxv1alpha1.ToolSpec{
						Type:          eirenyxv1alpha1.ToolTrivy,
						Enabled:       false,
						Namespace:     "trivy-system",
						InstallMethod: eirenyxv1alpha1.HelmInstall,
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			resource := &eirenyxv1alpha1.Tool{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance Tool")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &ToolReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
