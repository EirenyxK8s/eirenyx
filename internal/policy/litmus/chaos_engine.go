package litmus

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var ChaosEngineGVK = schema.GroupVersionKind{
	Group:   "litmuschaos.io",
	Version: "v1alpha1",
	Kind:    "ChaosEngine",
}

var ChaosEngineTypeMeta = v1.TypeMeta{
	APIVersion: ChaosEngineGVK.GroupVersion().String(),
	Kind:       ChaosEngineGVK.Kind,
}

type ChaosEngine struct {
	v1.TypeMeta   `json:",inline"`
	v1.ObjectMeta `json:"metadata,omitempty"`

	Spec ChaosEngineSpec `json:"spec"`
}

type ChaosEngineSpec struct {
	EngineState string            `json:"engineState"`
	AppInfo     ChaosAppInfo      `json:"appinfo"`
	Experiments []ChaosExperiment `json:"experiments"`
}

type ChaosAppInfo struct {
	AppNS    string `json:"appns"`
	AppLabel string `json:"applabel"`
	AppKind  string `json:"appkind"`
}

type ChaosExperiment struct {
	Name string              `json:"name"`
	Spec ChaosExperimentSpec `json:"spec"`
}

type ChaosExperimentSpec struct {
	Components ChaosComponents `json:"components"`
}

type ChaosComponents struct {
	Env []ChaosEnvVar `json:"env"`
}

type ChaosEnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func (c *ChaosEngine) ToUnstructured() (*unstructured.Unstructured, error) {
	objMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(c)
	if err != nil {
		return nil, err
	}

	u := &unstructured.Unstructured{Object: objMap}
	return u, nil
}
