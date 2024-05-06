/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1

import (
	"fmt"

	jsonpatch "github.com/evanphx/json-patch"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

// What follows is a list of all valid states for an Installation object.
const (
	InstallationStateWaiting                string = "Waiting"
	InstallationStateCopyingArtifacts       string = "CopyingArtifacts"
	InstallationStateEnqueued               string = "Enqueued"
	InstallationStateInstalling             string = "Installing"
	InstallationStateInstalled              string = "Installed"
	InstallationStateKubernetesInstalled    string = "KubernetesInstalled"
	InstallationStateAddonsInstalling       string = "AddonsInstalling"
	InstallationStateHelmChartUpdateFailure string = "HelmChartUpdateFailure"
	InstallationStateObsolete               string = "Obsolete"
	InstallationStateFailed                 string = "Failed"
	InstallationStateUnknown                string = "Unknown"
	InstallationStatePendingChartCreation   string = "PendingChartCreation"
)

// NodeStatus is used to keep track of the status of a cluster node, we
// only hold its name and a hash of the node's status. Whenever the node
// status change we will be able to capture it and update the hash.
type NodeStatus struct {
	Name string `json:"name"`
	Hash string `json:"hash"`
}

// ArtifactsLocation defines a location from where we can download an
// airgap bundle. It contains individual URLs for each component of the
// bundle. These URLs are expected to point to a registry running inside
// the cluster, authentication for the registry is read from the cluster
// at execution time so they do not need to be provided here.
type ArtifactsLocation struct {
	Images                  string `json:"images"`
	HelmCharts              string `json:"helmCharts"`
	EmbeddedClusterBinary   string `json:"embeddedClusterBinary"`
	EmbeddedClusterMetadata string `json:"embeddedClusterMetadata"`
}

// LicenseInfo holds information about the license used to install the cluster.
type LicenseInfo struct {
	IsSnapshotSupported bool `json:"isSnapshotSupported"`
}

// InstallationSpec defines the desired state of Installation.
type InstallationSpec struct {
	// ClusterID holds the cluster, generated during the installation.
	ClusterID string `json:"clusterID,omitempty"`
	// MetricsBaseURL holds the base URL for the metrics server.
	MetricsBaseURL string `json:"metricsBaseURL,omitempty"`
	// AirGap indicates if the installation is airgapped.
	AirGap bool `json:"airGap"`
	// Artifacts holds the location of the airgap bundle.
	Artifacts *ArtifactsLocation `json:"artifacts,omitempty"`
	// Config holds the configuration used at installation time.
	Config *ConfigSpec `json:"config,omitempty"`
	// EndUserK0sConfigOverrides holds the end user k0s config overrides
	// used at installation time.
	EndUserK0sConfigOverrides string `json:"endUserK0sConfigOverrides,omitempty"`
	// BinaryName holds the name of the binary used to install the cluster.
	// this will follow the pattern 'appslug-channelslug'
	BinaryName string `json:"binaryName,omitempty"`
	// LicenseInfo holds information about the license used to install the cluster.
	LicenseInfo *LicenseInfo `json:"licenseInfo,omitempty"`
	// UnknownConfigProperties is used when migrating between two different versions
	// of the CRD, eg: on v2 of the CRD a new field has been introduced, the cluster
	// does not know yet about this new field as it only knowns about the v1 version.
	// To avoid losing information the v2 to v1 diff is stored here as a JSON merge
	// patch marshaled as YAML. Look at this field as a patch that will be applied
	// on top of the Config field above before reconciling.
	UnknownConfigProperties string `json:"unknownConfigProperties,omitempty"`
}

// StoreUnknownConfigProperties creates a patch between .Config and the provided yaml
// string. Stores the patch on the UnknownConfigProperties field. If both are equal
// then the field is set to an empty string.
func (i *InstallationSpec) StoreUnknownConfigProperties(v2 string) error {
	v1yaml, err := yaml.Marshal(i.Config)
	if err != nil {
		return fmt.Errorf("failed to marshall current config to yaml: %w", err)
	}
	v1json, err := yaml.YAMLToJSON(v1yaml)
	if err != nil {
		return fmt.Errorf("failed to current config to json: %w", err)
	}
	v2json, err := yaml.YAMLToJSON([]byte(v2))
	if err != nil {
		return fmt.Errorf("failed to convert new config to json: %w", err)
	}
	if jsonpatch.Equal(v1json, v2json) {
		i.UnknownConfigProperties = ""
		return nil
	}
	patch, err := jsonpatch.CreateMergePatch(v1json, v2json)
	if err != nil {
		return fmt.Errorf("failed to create patch between configs: %w", err)
	}
	patchYAML, err := yaml.JSONToYAML(patch)
	if err != nil {
		return fmt.Errorf("failed to convert config patch to yaml: %w", err)
	}
	i.UnknownConfigProperties = string(patchYAML)
	return nil
}

// ApplyUnknownConfigProperties applies the unknown config properties to the
// Config field of the Installation object. This function may reset the content
// of the Config property to include the now known fields.
func (i *InstallationSpec) ApplyUnknownConfigProperties() error {
	if i.UnknownConfigProperties == "" {
		return nil
	}
	originalYAML, err := yaml.Marshal(i.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal original config: %w", err)
	}
	original, err := yaml.YAMLToJSON(originalYAML)
	if err != nil {
		return fmt.Errorf("failed to convert original config to json: %w", err)
	}
	patch, err := yaml.YAMLToJSON([]byte(i.UnknownConfigProperties))
	if err != nil {
		return fmt.Errorf("failed to convert patch to JSON: %w", err)
	}
	result, err := jsonpatch.MergePatch(original, patch)
	if err != nil {
		return fmt.Errorf("failed to apply patch: %w", err)
	}
	if jsonpatch.Equal(original, result) {
		return nil
	}
	asyaml, err := yaml.JSONToYAML(result)
	if err != nil {
		return fmt.Errorf("failed to convert JSON to YAML: %w", err)
	}
	var config ConfigSpec
	if err := yaml.Unmarshal(asyaml, &config); err != nil {
		return fmt.Errorf("failed to unmarshal patched config: %w", err)
	}
	i.Config = &config
	return nil
}

// InstallationStatus defines the observed state of Installation
type InstallationStatus struct {
	// NodesStatus is a list of nodes and their status.
	NodesStatus []NodeStatus `json:"nodesStatus,omitempty"`
	// State holds the current state of the installation.
	State string `json:"state,omitempty"`
	// Reason holds the reason for the current state.
	Reason string `json:"reason,omitempty"`
	// PendingCharts holds the list of charts that are being created or updated.
	PendingCharts []string `json:"pendingCharts,omitempty"`
}

// SetState sets the installation state and reason.
func (s *InstallationStatus) SetState(state string, reason string, pendingCharts []string) {
	s.State = state
	s.Reason = reason
	s.PendingCharts = pendingCharts
}

func (s *InstallationStatus) GetKubernetesInstalled() bool {
	if s.State == InstallationStateInstalled ||
		s.State == InstallationStateKubernetesInstalled ||
		s.State == InstallationStateAddonsInstalling ||
		s.State == InstallationStatePendingChartCreation ||
		s.State == InstallationStateHelmChartUpdateFailure {
		return true
	}
	return false
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster
//+kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state",description="State of the installation"
//+kubebuilder:printcolumn:name="InstallerVersion",type="string",JSONPath=".spec.config.version",description="Installer version"
//+kubebuilder:printcolumn:name="CreatedAt",type="string",JSONPath=".metadata.creationTimestamp",description="Creation time of the installation"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="Age of the resource"

// Installation is the Schema for the installations API
type Installation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InstallationSpec   `json:"spec,omitempty"`
	Status InstallationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// InstallationList contains a list of Installation
type InstallationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Installation `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Installation{}, &InstallationList{})
}
