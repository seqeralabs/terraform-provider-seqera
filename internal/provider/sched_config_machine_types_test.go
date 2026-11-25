package provider

import (
	"encoding/json"
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
		{name: "listed", body: `{"machineTypes":["m5.large"]}`, want: []string{"m5.large"}},
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
			if len(sc.MachineTypes) != len(tc.want) || (len(tc.want) > 0 && sc.MachineTypes[0] != tc.want[0]) {
				t.Fatalf("machineTypes = %v, want %v", sc.MachineTypes, tc.want)
			}
		})
	}

	// Every cloud compute environment shares SchedConfig, so the transform
	// must apply through each of their config types.
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
	}
	for cloud, decode := range clouds {
		t.Run(cloud+" null machineTypes reads as empty", func(t *testing.T) {
			sc, err := decode([]byte(`{"region":"r","schedConfig":{"machineTypes":null}}`))
			if err != nil {
				t.Fatal(err)
			}
			if sc == nil || sc.MachineTypes == nil || len(sc.MachineTypes) != 0 {
				t.Fatalf("schedConfig = %+v, want machineTypes []", sc)
			}
		})
		t.Run(cloud+" null schedConfig stays null", func(t *testing.T) {
			sc, err := decode([]byte(`{"region":"r","schedConfig":null}`))
			if err != nil {
				t.Fatal(err)
			}
			if sc != nil {
				t.Fatalf("schedConfig = %+v, want nil", sc)
			}
		})
	}
}
