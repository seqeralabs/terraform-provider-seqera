package provider

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/seqeralabs/terraform-provider-seqera/internal/sdk/models/shared"
)

// The API returns an unset machineTypes as null, and the provider drops an
// empty list from requests. The overlay transform reads null back as [] so a
// configured `machine_types = []` does not force a replacement on every plan
// (#321). A null schedConfig (Classic mode) must stay null.
func TestSchedConfigMachineTypesNullReadsAsEmpty(t *testing.T) {
	cases := []struct {
		name string
		body string
		want []string
	}{
		{name: "null", body: `{"machineTypes":null}`, want: []string{}},
		{name: "absent", body: `{}`, want: []string{}},
		{name: "empty", body: `{"machineTypes":[]}`, want: []string{}},
		{name: "listed", body: `{"machineTypes":["m5.large","c5.xlarge"]}`, want: []string{"m5.large", "c5.xlarge"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var sc shared.SchedConfig
			if err := json.Unmarshal([]byte(tc.body), &sc); err != nil {
				t.Fatal(err)
			}
			if sc.MachineTypes == nil {
				t.Fatalf("machineTypes is nil, want %v", tc.want)
			}
			if !slices.Equal(sc.MachineTypes, tc.want) {
				t.Fatalf("machineTypes = %v, want %v", sc.MachineTypes, tc.want)
			}
		})
	}

	// Every compute environment with Intelligent Compute shares SchedConfig, so
	// the transform must apply through each of their config types: the
	// per-cloud resources (seqera_*_cloud_ce) and the members of the
	// ComputeConfig union used by seqera_compute_env.
	union := func(discriminator string, get func(shared.ComputeConfig) *shared.SchedConfig) func([]byte) (*shared.SchedConfig, error) {
		return func(b []byte) (*shared.SchedConfig, error) {
			var m map[string]any
			if err := json.Unmarshal(b, &m); err != nil {
				return nil, err
			}
			m["discriminator"] = discriminator
			b, _ = json.Marshal(m)
			var cc shared.ComputeConfig
			if err := json.Unmarshal(b, &cc); err != nil {
				return nil, err
			}
			return get(cc), nil
		}
	}
	clouds := map[string]func(body []byte) (*shared.SchedConfig, error){
		"aws": func(b []byte) (*shared.SchedConfig, error) {
			var c shared.AwsCloudConfig
			err := json.Unmarshal(b, &c)
			return c.IntelligentComputeConfig, err
		},
		"azure": func(b []byte) (*shared.SchedConfig, error) {
			var c shared.AzCloudConfig
			err := json.Unmarshal(b, &c)
			return c.IntelligentComputeConfig, err
		},
		"google": func(b []byte) (*shared.SchedConfig, error) {
			var c shared.GoogleCloudConfig
			err := json.Unmarshal(b, &c)
			return c.IntelligentComputeConfig, err
		},
		"compute_env aws_cloud": union("aws-cloud", func(cc shared.ComputeConfig) *shared.SchedConfig {
			return cc.AWSCloudConfiguration.IntelligentComputeConfig
		}),
		"compute_env azure_cloud": union("azure-cloud", func(cc shared.ComputeConfig) *shared.SchedConfig {
			return cc.AzureCloudConfiguration.IntelligentComputeConfig
		}),
		"compute_env google_cloud": union("google-cloud", func(cc shared.ComputeConfig) *shared.SchedConfig {
			return cc.GoogleCloudConfiguration.IntelligentComputeConfig
		}),
		"compute_env local_platform": union("local-platform", func(cc shared.ComputeConfig) *shared.SchedConfig {
			return cc.LocalExecutionConfiguration.IntelligentComputeConfig
		}),
	}
	for cloud, decode := range clouds {
		t.Run(cloud+" null machineTypes reads as empty", func(t *testing.T) {
			sc, err := decode([]byte(`{"region":"r","workDir":"s3://b","schedConfig":{"machineTypes":null}}`))
			if err != nil {
				t.Fatal(err)
			}
			if sc == nil || sc.MachineTypes == nil || len(sc.MachineTypes) != 0 {
				t.Fatalf("schedConfig = %+v, want machineTypes []", sc)
			}
		})
		t.Run(cloud+" null schedConfig stays null", func(t *testing.T) {
			sc, err := decode([]byte(`{"region":"r","workDir":"s3://b","schedConfig":null}`))
			if err != nil {
				t.Fatal(err)
			}
			if sc != nil {
				t.Fatalf("schedConfig = %+v, want nil", sc)
			}
		})
	}
}
