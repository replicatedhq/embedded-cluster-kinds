package v1beta1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// these types were copied from https://github.com/k0sproject/k0s/blob/7a7255e/pkg/apis/k0s/v1beta1/extensions.go

// HelmExtensions specifies settings for cluster helm based extensions
type HelmExtensions struct {
	ConcurrencyLevel int                  `json:"concurrencyLevel"`
	Repositories     RepositoriesSettings `json:"repositories"`
	Charts           ChartsSettings       `json:"charts"`
}

// RepositoriesSettings repository settings
type RepositoriesSettings []Repository

// Repository describes single repository entry. Fields map to the CLI flags for the "helm add" command
type Repository struct {
	Name     string `json:"name"`
	URL      string `json:"url"`
	CAFile   string `json:"caFile"`
	CertFile string `json:"certFile"`
	Insecure bool   `json:"insecure"`
	KeyFile  string `json:"keyfile"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// ChartsSettings charts settings
type ChartsSettings []Chart

// Chart single helm addon
type Chart struct {
	Name      string `json:"name"`
	ChartName string `json:"chartname"`
	Version   string `json:"version"`
	Values    string `json:"values"`
	TargetNS  string `json:"namespace"`
	// Timeout specifies the timeout for how long to wait for the chart installation to finish.
	// A duration string is a sequence of decimal numbers, each with optional fraction and a unit suffix, such as "300ms" or "2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	Timeout metav1.Duration `json:"timeout"`
	Order   int             `json:"order"`
}
