package types

import (
	"encoding/json"
	"fmt"
	"time"

	k0sv1beta1 "github.com/k0sproject/k0s/pkg/apis/k0s/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// ClusterConfig is an alias of k0sv1beta1.ClusterConfig
type ClusterConfig k0sv1beta1.ClusterConfig

// UnmarshalJSON implements the json.Unmarshaler interface.
func (c *ClusterConfig) UnmarshalJSON(data []byte) error {
	var u unstructured.Unstructured
	if err := json.Unmarshal(data, &u); err != nil {
		return err
	}

	spec, found, err := unstructured.NestedMap(u.Object, "spec")
	if err != nil {
		return fmt.Errorf("error accessing spec: %v", err)
	}
	if !found {
		return fmt.Errorf("spec not found in unstructured data")
	}

	extensions, found, err := unstructured.NestedMap(spec, "extensions")
	if err != nil {
		return fmt.Errorf("error accessing extensions: %v", err)
	}
	if !found {
		return fmt.Errorf("extensions not found in unstructured data")
	}

	helm, found, err := unstructured.NestedMap(extensions, "helm")
	if err != nil {
		return fmt.Errorf("error accessing helm extensions: %v", err)
	}
	if !found {
		return fmt.Errorf("helm extensions not found in unstructured data")
	}

	charts, found, err := unstructured.NestedSlice(helm, "charts")
	if err != nil {
		return fmt.Errorf("error accessing helm charts: %v", err)
	}
	if !found {
		return fmt.Errorf("helm charts not found in unstructured data")
	}

	for i, chart := range charts {
		chartMap, ok := chart.(map[string]interface{})
		if !ok {
			continue
		}
		timeout, exists := chartMap["timeout"]
		if !exists {
			continue
		}
		intTimeout, ok := timeout.(int64)
		if !ok {
			continue
		}
		chartMap["timeout"] = metav1.Duration{Duration: time.Duration(intTimeout)}.ToUnstructured()
		charts[i] = chartMap
	}

	if err := unstructured.SetNestedSlice(helm, charts, "charts"); err != nil {
		return fmt.Errorf("error setting updated helm charts: %v", err)
	}
	if err := unstructured.SetNestedMap(extensions, helm, "helm"); err != nil {
		return fmt.Errorf("error setting updated helm extensions: %v", err)
	}
	if err := unstructured.SetNestedMap(spec, extensions, "extensions"); err != nil {
		return fmt.Errorf("error setting updated extensions: %v", err)
	}
	if err := unstructured.SetNestedMap(u.Object, spec, "spec"); err != nil {
		return fmt.Errorf("error setting updated spec: %v", err)
	}

	var k0sConfig k0sv1beta1.ClusterConfig
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &k0sConfig); err != nil {
		return fmt.Errorf("failed to convert unstructured to ClusterConfig: %v", err)
	}

	*c = ClusterConfig(k0sConfig)
	return nil
}
