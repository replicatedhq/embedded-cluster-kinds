package types

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestClusterConfig_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantTimeout metav1.Duration
		wantErr     bool
	}{
		{
			name: "string timeout",
			input: `{
				"apiVersion": "k0s.k0sproject.io/v1beta1",
				"kind": "ClusterConfig",
				"metadata": {
					"name": "test-cluster"
				},
				"spec": {
					"extensions": {
						"helm": {
							"charts": [
								{
									"name": "test-chart",
									"timeout": "5m"
								}
							]
						}
					}
				}
			}`,
			wantTimeout: metav1.Duration{Duration: 5 * time.Minute},
			wantErr:     false,
		}, {
			name: "int timeout",
			input: `{
				"apiVersion": "k0s.k0sproject.io/v1beta1",
				"kind": "ClusterConfig",
				"metadata": {
					"name": "test-cluster"
				},
				"spec": {
					"extensions": {
						"helm": {
							"charts": [
								{
									"name": "test-chart",
									"timeout": 300000000000
								}
							]
						}
					}
				}
			}`,
			wantTimeout: metav1.Duration{Duration: 5 * time.Minute},
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got ClusterConfig
			err := json.Unmarshal([]byte(tt.input), &got)
			if (err != nil) != tt.wantErr {
				t.Errorf("ClusterConfig.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, tt.wantTimeout, got.Spec.Extensions.Helm.Charts[0].Timeout)
		})
	}
}
