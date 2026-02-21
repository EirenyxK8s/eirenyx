package report

import (
	"context"
	"encoding/json"
	"math/rand"
	"time"

	"github.com/EirenyxK8s/eirenyx/internal/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logger "sigs.k8s.io/controller-runtime/pkg/log"
)

var excludedNamespaces = []string{
	"kube-system",
	"eirenyx-system",
	"falco",
	"litmus",
	"trivy-system",
}

// PodDetails holds information about a pod that will be returned as JSON
type PodDetails struct {
	PodName    string   `json:"podName"`
	Namespace  string   `json:"namespace"`
	Containers []string `json:"containers"`
}

// GetPodDetails fetches unique random pods from all namespaces excluding the provided list
func GetPodDetails(ctx context.Context, k8sClient client.Client, podCount int) runtime.RawExtension {
	log := logger.FromContext(ctx)

	namespaces := &corev1.NamespaceList{}
	if err := k8sClient.List(ctx, namespaces); err != nil {
		log.Error(err, "Failed to list namespaces")
		return runtime.RawExtension{}
	}

	var includedNamespaces []string
	for _, ns := range namespaces.Items {
		if !utils.Contains(excludedNamespaces, ns.Name) {
			includedNamespaces = append(includedNamespaces, ns.Name)
		}
	}

	var collectedPods []corev1.Pod
	for _, namespace := range includedNamespaces {
		podList := &corev1.PodList{}
		if err := k8sClient.List(ctx, podList, client.InNamespace(namespace)); err != nil {
			log.Error(err, "Failed to list pods in namespace", "namespace", namespace)
			return runtime.RawExtension{}
		}
		collectedPods = append(collectedPods, podList.Items...)
	}

	if len(collectedPods) == 0 {
		log.Info("No pods found in the selected namespaces")
		return runtime.RawExtension{
			Raw: []byte(`{"message": "No pods found."}`),
		}
	}

	rand.New(rand.NewSource(time.Now().UnixNano()))
	rand.Shuffle(len(collectedPods), func(i, j int) {
		collectedPods[i], collectedPods[j] = collectedPods[j], collectedPods[i]
	})

	podSet := make(map[string]PodDetails, podCount)
	for _, p := range collectedPods {
		if len(podSet) >= podCount {
			break
		}

		key := p.Namespace + "/" + p.Name
		if _, exists := podSet[key]; !exists {
			podSet[key] = PodDetails{
				PodName:    p.Name,
				Namespace:  p.Namespace,
				Containers: ExtractContainerNames(p),
			}
		}
	}

	podDetails := make([]PodDetails, 0, len(podSet))
	for _, v := range podSet {
		podDetails = append(podDetails, v)
	}

	podDetailsJSON, err := json.Marshal(podDetails)
	if err != nil {
		log.Error(err, "Failed to marshal pod details to JSON")
		return runtime.RawExtension{}
	}

	return runtime.RawExtension{Raw: podDetailsJSON}
}

// ExtractContainerNames extracts the container names from a pod
func ExtractContainerNames(pod corev1.Pod) []string {
	containerNames := make([]string, 0, len(pod.Spec.Containers))
	for _, container := range pod.Spec.Containers {
		containerNames = append(containerNames, container.Name)
	}
	return containerNames
}
