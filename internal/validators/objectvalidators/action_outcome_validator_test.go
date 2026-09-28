package objectvalidators

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// outcomeBlock is the state of one of the launch / agent / pipeline blocks.
type outcomeBlock int

const (
	blockNull outcomeBlock = iota
	blockSet
	blockUnknown
)

// makeActionOutcomeRequest builds the request the validator sees when attached
// to seqera_action's `launch` attribute. Each block has a single string
// attribute, which is all the validator needs. A nil responseType is null (the
// API then defaults to "pipeline"); unknownType makes it unknown.
func makeActionOutcomeRequest(t *testing.T, responseType *string, unknownType bool, launch, agent, pipeline outcomeBlock) validator.ObjectRequest {
	t.Helper()

	block := func(field string) (tftypes.Object, resourceschema.SingleNestedAttribute, map[string]attr.Type) {
		return tftypes.Object{AttributeTypes: map[string]tftypes.Type{field: tftypes.String}},
			resourceschema.SingleNestedAttribute{Optional: true, Attributes: map[string]resourceschema.Attribute{field: resourceschema.StringAttribute{Optional: true}}},
			map[string]attr.Type{field: types.StringType}
	}
	launchType, launchAttr, launchAttrTypes := block("pipeline")
	agentType, agentAttr, _ := block("agent_config_id")
	pipelineType, pipelineAttr, _ := block("target_pipeline_version_id")

	raw := func(typ tftypes.Object, field string, state outcomeBlock) tftypes.Value {
		switch state {
		case blockSet:
			return tftypes.NewValue(typ, map[string]tftypes.Value{field: tftypes.NewValue(tftypes.String, "x")})
		case blockUnknown:
			return tftypes.NewValue(typ, tftypes.UnknownValue)
		default:
			return tftypes.NewValue(typ, nil)
		}
	}

	typeRaw := tftypes.NewValue(tftypes.String, nil)
	switch {
	case unknownType:
		typeRaw = tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
	case responseType != nil:
		typeRaw = tftypes.NewValue(tftypes.String, *responseType)
	}

	rootType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"response_type": tftypes.String, "launch": launchType, "agent": agentType, "pipeline": pipelineType,
	}}
	rootRaw := tftypes.NewValue(rootType, map[string]tftypes.Value{
		"response_type": typeRaw,
		"launch":        raw(launchType, "pipeline", launch),
		"agent":         raw(agentType, "agent_config_id", agent),
		"pipeline":      raw(pipelineType, "target_pipeline_version_id", pipeline),
	})
	schema := resourceschema.Schema{Attributes: map[string]resourceschema.Attribute{
		"response_type": resourceschema.StringAttribute{Optional: true},
		"launch":        launchAttr, "agent": agentAttr, "pipeline": pipelineAttr,
	}}

	launchValue := types.ObjectNull(launchAttrTypes)
	switch launch {
	case blockSet:
		launchValue = types.ObjectValueMust(launchAttrTypes, map[string]attr.Value{"pipeline": types.StringValue("x")})
	case blockUnknown:
		launchValue = types.ObjectUnknown(launchAttrTypes)
	}

	return validator.ObjectRequest{
		Path:        path.Root("launch"),
		ConfigValue: launchValue,
		Config:      tfsdk.Config{Schema: schema, Raw: rootRaw},
	}
}

func TestActionOutcomeValidator(t *testing.T) {
	agent, pipeline := "agent", "pipeline"
	cases := []struct {
		name         string
		responseType *string
		unknownType  bool
		launch       outcomeBlock
		agent        outcomeBlock
		pipeline     outcomeBlock
		wantErrPath  string // empty means valid
	}{
		{name: "default response with launch", launch: blockSet},
		{name: "pipeline response with launch and target pipeline", responseType: &pipeline, launch: blockSet, pipeline: blockSet},
		{name: "default response without launch", wantErrPath: "launch"},
		{name: "pipeline response without launch", responseType: &pipeline, wantErrPath: "launch"},
		{name: "pipeline response with agent", responseType: &pipeline, launch: blockSet, agent: blockSet, wantErrPath: "agent"},
		{name: "agent response with agent only", responseType: &agent, agent: blockSet},
		{name: "agent response with launch", responseType: &agent, agent: blockSet, launch: blockSet, wantErrPath: "launch"},
		{name: "agent response with target pipeline", responseType: &agent, agent: blockSet, pipeline: blockSet, wantErrPath: "pipeline"},
		{name: "agent response without agent", responseType: &agent, wantErrPath: "agent"},
		{name: "unknown response type is not judged", unknownType: true},
		{name: "unknown launch is not judged", responseType: &agent, agent: blockSet, launch: blockUnknown},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := makeActionOutcomeRequest(t, tc.responseType, tc.unknownType, tc.launch, tc.agent, tc.pipeline)
			resp := &validator.ObjectResponse{}
			ActionOutcomeValidator().ValidateObject(context.Background(), req, resp)

			if tc.wantErrPath == "" {
				if resp.Diagnostics.HasError() {
					t.Fatalf("expected no error, got %v", resp.Diagnostics)
				}
				return
			}
			if !resp.Diagnostics.HasError() {
				t.Fatalf("expected an error on %s, got none", tc.wantErrPath)
			}
			for _, d := range resp.Diagnostics.Errors() {
				if dp, ok := d.(interface{ Path() path.Path }); ok && dp.Path().Equal(path.Root(tc.wantErrPath)) {
					return
				}
			}
			t.Fatalf("expected an error on %s, got %v", tc.wantErrPath, resp.Diagnostics)
		})
	}
}
