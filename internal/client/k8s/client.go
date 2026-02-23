package k8s

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Client wraps the Kubernetes clientset with domain-specific helpers.
type Client struct {
	k8s kubernetes.Interface
}

// NewClient constructs a Client from any kubernetes.Interface.
func NewClient(k8s kubernetes.Interface) *Client {
	return &Client{k8s: k8s}
}

// NewClientFromConfig builds a real in-cluster or kubeconfig-based Client.
func NewClientFromConfig(cfg *rest.Config) (*Client, error) {
	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("creating kubernetes clientset: %w", err)
	}
	return NewClient(clientset), nil
}

// EnsureNamespace creates the namespace if it does not already exist.
func (c *Client) EnsureNamespace(ctx context.Context, namespace string) error {
	_, err := c.k8s.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err == nil {
		return nil
	}
	if !apierrors.IsNotFound(err) {
		return fmt.Errorf("getting namespace %s: %w", namespace, err)
	}

	_, err = c.k8s.CoreV1().Namespaces().Create(ctx, &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: namespace},
	}, metav1.CreateOptions{})

	if err != nil && !apierrors.IsAlreadyExists(err) {
		return fmt.Errorf("creating namespace %s: %w", namespace, err)
	}
	return nil
}

// DeleteNamespace deletes the namespace. Idempotent — not found is not an error.
func (c *Client) DeleteNamespace(ctx context.Context, ns string) error {
	err := c.k8s.CoreV1().Namespaces().Delete(ctx, ns, metav1.DeleteOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("deleting namespace %s: %w", ns, err)
	}
	return nil
}

// IsDeploymentReady reports whether the named deployment is fully available.
func (c *Client) IsDeploymentReady(ctx context.Context, namespace, name string) (bool, error) {
	deploy, err := c.k8s.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("getting deployment %s/%s: %w", namespace, name, err)
	}
	return isDeploymentReady(deploy), nil
}

// IsDaemonSetReady reports whether all desired pods of the named DaemonSet are ready.
func (c *Client) IsDaemonSetReady(ctx context.Context, namespace, name string) (bool, error) {
	ds, err := c.k8s.AppsV1().DaemonSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("getting daemonset %s/%s: %w", namespace, name, err)
	}
	return isDaemonSetReady(ds), nil
}

func isDeploymentReady(d *appsv1.Deployment) bool {
	if d.Status.ObservedGeneration < d.Generation {
		return false
	}
	for _, cond := range d.Status.Conditions {
		if cond.Type == appsv1.DeploymentAvailable && cond.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func isDaemonSetReady(ds *appsv1.DaemonSet) bool {
	if ds.Status.ObservedGeneration < ds.Generation {
		return false
	}
	return ds.Status.NumberReady > 0 &&
		ds.Status.NumberReady >= ds.Status.DesiredNumberScheduled
}
