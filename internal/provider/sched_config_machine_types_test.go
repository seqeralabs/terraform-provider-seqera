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

	t.Run("null schedConfig stays null", func(t *testing.T) {
		var cfg shared.AwsCloudConfig
		if err := json.Unmarshal([]byte(`{"region":"eu-west-1","schedConfig":null}`), &cfg); err != nil {
			t.Fatal(err)
		}
		if cfg.IntelligentComputeConfig != nil {
			t.Fatalf("schedConfig = %+v, want nil", cfg.IntelligentComputeConfig)
		}
	})
}
