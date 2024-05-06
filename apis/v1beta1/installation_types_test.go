package v1beta1

import (
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"
)

func TestStoreUnknownConfigProperties(t *testing.T) {
	type test struct {
		Name          string                 `yaml:"name"`
		ConfigSpec    ConfigSpec             `yaml:"configSpec"`
		NewConfigSpec string                 `yaml:"newConfigSpec"`
		Expected      map[string]interface{} `yaml:"expected"`
	}

	for tname, tt := range parseTestsYAML[test](t, "store-unknown-properties-") {
		t.Run(tname, func(t *testing.T) {
			in := InstallationSpec{Config: &tt.ConfigSpec}
			err := in.StoreUnknownConfigProperties(tt.NewConfigSpec)
			require.NoError(t, err)
			result := map[string]interface{}{}
			err = yaml.Unmarshal([]byte(in.UnknownConfigProperties), &result)
			require.NoError(t, err)
			require.Equal(t, tt.Expected, result, "unexpected unknown config properties")
		})
	}
}

func TestApplyUnknownConfigProperties(t *testing.T) {
	type test struct {
		Name                    string      `yaml:"name"`
		ConfigSpec              *ConfigSpec `yaml:"configSpec"`
		UnknownConfigProperties string      `yaml:"unknownConfigProperties"`
		Expected                *ConfigSpec `yaml:"expected"`
	}

	for tname, tt := range parseTestsYAML[test](t, "apply-unknown-properties-") {
		t.Run(tname, func(t *testing.T) {
			in := InstallationSpec{
				Config:                  tt.ConfigSpec,
				UnknownConfigProperties: tt.UnknownConfigProperties,
			}
			err := in.ApplyUnknownConfigProperties()
			require.NoError(t, err)
			require.Equal(t, tt.Expected, in.Config, "unexpected resulting config")
		})
	}
}
