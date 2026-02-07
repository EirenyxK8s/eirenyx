package k8s

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	kubeClient *kubernetes.Clientset
	initOnce   sync.Once
	initErr    error
)

// GetK8sClient returns a singleton Kubernetes clientset
func GetK8sClient() (*kubernetes.Clientset, error) {
	initOnce.Do(func() {
		cfg := ctrl.GetConfigOrDie()
		kubeClient, initErr = kubernetes.NewForConfig(cfg)
		if initErr != nil {
			initErr = errors.New(fmt.Sprintf("failed to create kube client: %s", initErr))
		}
	})

	return kubeClient, initErr
}

// EnsureK8sNamespace ensures that the specified namespace exists in the cluster
func EnsureK8sNamespace(ctx context.Context, namespace string) error {
	k8sClient, err := GetK8sClient()
	if err != nil {
		return err
	}

	_, err = k8sClient.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = k8sClient.CoreV1().Namespaces().Create(
			ctx,
			&corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: namespace,
				},
			},
			metav1.CreateOptions{},
		)
		if err != nil {
			return errors.New(fmt.Sprintf("failed to create namespace %s: %s", namespace, err))
		}
		return nil
	}

	if err != nil {
		return errors.New(fmt.Sprintf("failed to get namespace %s: %s", namespace, err))
	}

	return nil
}

// EnsureNamespaceDeleted deletes the specified namespace
func EnsureNamespaceDeleted(ctx context.Context, ns string) error {
	k8sClient, err := GetK8sClient()
	if err != nil {
		return err
	}

	if err = k8sClient.CoreV1().Namespaces().Delete(ctx, ns, metav1.DeleteOptions{}); err != nil {
		return errors.Wrap(err, "failed to delete namespace")
	}

	return nil
}

// GetDeployment retrieves a deployment by namespace and name
func GetDeployment(ctx context.Context, namespace, name string) (*appsv1.Deployment, error) {
	k8sClient, err := GetK8sClient()
	if err != nil {
		return nil, err
	}

	return k8sClient.AppsV1().
		Deployments(namespace).
		Get(ctx, name, metav1.GetOptions{})
}

// EnsureDeploymentRun waits for the deployment to become ready
func EnsureDeploymentRun(ctx context.Context, namespace string, deployName string) bool {
	log := logf.FromContext(ctx)
	k8sClient, err := GetK8sClient()
	if err != nil {
		log.Error(err, "failed to get k8s client")
		return false
	}

	for start := time.Now(); time.Since(start) < 3*time.Minute; {
		deploy, err := k8sClient.AppsV1().
			Deployments(namespace).
			Get(ctx, deployName, metav1.GetOptions{})

		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}

		if isDeploymentReady(deploy) {
			return true
		}

		time.Sleep(5 * time.Second)
	}
	return false
}

func isDeploymentReady(d *appsv1.Deployment) bool {
	if d.Status.ObservedGeneration < d.Generation {
		return false
	}
	for _, cond := range d.Status.Conditions {
		if cond.Type == appsv1.DeploymentAvailable &&
			cond.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func IsDaemonSetReady(ctx context.Context, namespace, name string) bool {
	log := logf.FromContext(ctx)
	k8sClient, err := GetK8sClient()
	if err != nil {
		log.Error(err, "failed to get k8s client")
		return false
	}

	ds, err := k8sClient.AppsV1().
		DaemonSets(namespace).
		Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return false
	}

	if ds.Status.NumberReady == 0 {
		return false
	}

	if ds.Status.NumberReady < ds.Status.DesiredNumberScheduled {
		return false
	}

	return true
}
