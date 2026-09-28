package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// planFixture holds the three views the framework hands to plan modifiers.
type planFixture struct {
	state  tfsdk.State
	plan   tfsdk.Plan
	config tfsdk.Config
}

// objectValue builds a value of the schema's object type; attributes not in
// vals are null.
func objectValue(t *testing.T, s schema.Schema, vals map[string]tftypes.Value) tftypes.Value {
	t.Helper()
	typ := s.Type().TerraformType(context.Background()).(tftypes.Object)
	m := make(map[string]tftypes.Value, len(typ.AttributeTypes))
	for name, at := range typ.AttributeTypes {
		if v, ok := vals[name]; ok {
			m[name] = v
		} else {
			m[name] = tftypes.NewValue(at, nil)
		}
	}
	return tftypes.NewValue(typ, m)
}

func newPlanFixture(t *testing.T, s schema.Schema, state, plan, config map[string]tftypes.Value) planFixture {
	t.Helper()
	f := planFixture{
		plan:   tfsdk.Plan{Schema: s, Raw: objectValue(t, s, plan)},
		config: tfsdk.Config{Schema: s, Raw: objectValue(t, s, config)},
		state:  tfsdk.State{Schema: s, Raw: objectValue(t, s, nil)},
	}
	if state != nil {
		f.state.Raw = objectValue(t, s, state)
	}
	return f
}

// plannedString runs every plan modifier of a string attribute, as the
// framework does, and returns the resulting planned value.
func plannedString(t *testing.T, s schema.Schema, f planFixture, attr string) types.String {
	t.Helper()
	ctx := context.Background()
	a, ok := s.Attributes[attr].(schema.StringAttribute)
	if !ok {
		t.Fatalf("%s is not a string attribute", attr)
	}
	var sv, pv, cv types.String
	f.state.GetAttribute(ctx, path.Root(attr), &sv)
	f.plan.GetAttribute(ctx, path.Root(attr), &pv)
	f.config.GetAttribute(ctx, path.Root(attr), &cv)
	resp := planmodifier.StringResponse{PlanValue: pv}
	for _, m := range a.PlanModifiers {
		m.PlanModifyString(ctx, planmodifier.StringRequest{
			Path: path.Root(attr), State: f.state, Plan: f.plan, Config: f.config,
			StateValue: sv, PlanValue: resp.PlanValue, ConfigValue: cv,
		}, &resp)
	}
	return resp.PlanValue
}

func plannedInt64(t *testing.T, s schema.Schema, f planFixture, attr string) types.Int64 {
	t.Helper()
	ctx := context.Background()
	a, ok := s.Attributes[attr].(schema.Int64Attribute)
	if !ok {
		t.Fatalf("%s is not an int64 attribute", attr)
	}
	var sv, pv, cv types.Int64
	f.state.GetAttribute(ctx, path.Root(attr), &sv)
	f.plan.GetAttribute(ctx, path.Root(attr), &pv)
	f.config.GetAttribute(ctx, path.Root(attr), &cv)
	resp := planmodifier.Int64Response{PlanValue: pv}
	for _, m := range a.PlanModifiers {
		m.PlanModifyInt64(ctx, planmodifier.Int64Request{
			Path: path.Root(attr), State: f.state, Plan: f.plan, Config: f.config,
			StateValue: sv, PlanValue: resp.PlanValue, ConfigValue: cv,
		}, &resp)
	}
	return resp.PlanValue
}

var unknownString = tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
var unknownNumber = tftypes.NewValue(tftypes.Number, tftypes.UnknownValue)

func str(v string) tftypes.Value { return tftypes.NewValue(tftypes.String, v) }
func num(v int64) tftypes.Value  { return tftypes.NewValue(tftypes.Number, v) }

// agentUpdate models an in-place update of an existing agent: the user edits
// agent_instructions and binds serviceAccountID; every computed attribute
// arrives unknown in the proposed plan, as the framework marks it.
func agentUpdate(t *testing.T, s schema.Schema, serviceAccountID int64) planFixture {
	state := map[string]tftypes.Value{
		"workspace_id": num(5), "id": str("a1"), "name": str("triage"),
		"agent_instructions": str("old"), "service_account_id": num(10),
		"service_account_name": str("run-triage"), "status": str("active"),
		"date_created": str("2026-09-01T10:00:00Z"), "last_updated": str("2026-09-01T10:00:00Z"),
	}
	config := map[string]tftypes.Value{
		"workspace_id": num(5), "name": str("triage"), "agent_instructions": str("new"),
		"service_account_id": num(serviceAccountID),
	}
	plan := map[string]tftypes.Value{
		"workspace_id": num(5), "name": str("triage"), "agent_instructions": str("new"),
		"service_account_id": num(serviceAccountID),
		"id":                 unknownString, "service_account_name": unknownString, "status": unknownString,
		"date_created": unknownString, "last_updated": unknownString,
	}
	return newPlanFixture(t, s, state, plan, config)
}

// Attributes the API never changes after creation must not show as
// "(known after apply)" when another field is edited.
func TestAgentStableComputedAttributesKeepStateOnUpdate(t *testing.T) {
	s := resourceSchema(t, NewAgentResource())
	f := agentUpdate(t, s, 10)
	want := map[string]string{"id": "a1", "date_created": "2026-09-01T10:00:00Z", "status": "active"}
	for attr, v := range want {
		if got := plannedString(t, s, f, attr); got.IsUnknown() || got.ValueString() != v {
			t.Errorf("%s planned as %v, want %q from state", attr, got, v)
		}
	}
}

// service_account_name only changes when the agent is bound to another
// service account, so it keeps its state value unless service_account_id
// changes, and is unknown when it does.
func TestAgentServiceAccountNameFollowsServiceAccountID(t *testing.T) {
	s := resourceSchema(t, NewAgentResource())

	if got := plannedString(t, s, agentUpdate(t, s, 10), "service_account_name"); got.IsUnknown() || got.ValueString() != "run-triage" {
		t.Errorf("same service account: planned %v, want %q from state", got, "run-triage")
	}
	if got := plannedString(t, s, agentUpdate(t, s, 11), "service_account_name"); !got.IsUnknown() {
		t.Errorf("rebound service account: planned %v, want unknown", got)
	}

	create := newPlanFixture(t, s, nil,
		map[string]tftypes.Value{"workspace_id": num(5), "name": str("triage"), "agent_instructions": str("x"), "service_account_id": num(10), "service_account_name": unknownString},
		map[string]tftypes.Value{"workspace_id": num(5), "name": str("triage"), "agent_instructions": str("x"), "service_account_id": num(10)})
	if got := plannedString(t, s, create, "service_account_name"); !got.IsUnknown() {
		t.Errorf("create: planned %v, want unknown", got)
	}
}

func TestServiceAccountStableComputedAttributesKeepStateOnUpdate(t *testing.T) {
	s := resourceSchema(t, NewServiceAccountResource())
	state := map[string]tftypes.Value{
		"org_id": num(7), "id": num(42), "member_id": num(99), "name": str("ci"),
		"description": str("CI identity"), "created_at": str("2026-09-01T10:00:00Z"),
	}
	config := map[string]tftypes.Value{"org_id": num(7), "name": str("ci-agent")}
	plan := map[string]tftypes.Value{
		"org_id": num(7), "name": str("ci-agent"),
		"id": unknownNumber, "member_id": unknownNumber, "description": str(""), "created_at": unknownString,
	}
	f := newPlanFixture(t, s, state, plan, config)

	for attr, v := range map[string]int64{"id": 42, "member_id": 99} {
		if got := plannedInt64(t, s, f, attr); got.IsUnknown() || got.ValueInt64() != v {
			t.Errorf("%s planned as %v, want %d from state", attr, got, v)
		}
	}
	for attr, v := range map[string]string{"created_at": "2026-09-01T10:00:00Z"} {
		if got := plannedString(t, s, f, attr); got.IsUnknown() || got.ValueString() != v {
			t.Errorf("%s planned as %v, want %q from state", attr, got, v)
		}
	}
}

// Removing description from the configuration must clear it. The update is
// a PATCH that ignores a missing value but writes an empty string, so an
// unset description plans (and sends) "".
func TestServiceAccountDescriptionDefaultsToEmpty(t *testing.T) {
	s := resourceSchema(t, NewServiceAccountResource())
	a, ok := s.Attributes["description"].(schema.StringAttribute)
	if !ok {
		t.Fatal("description is not a string attribute")
	}
	if !a.Optional || !a.Computed || a.Default == nil {
		t.Fatalf("description: optional=%v computed=%v default=%v, want optional, computed and a default", a.Optional, a.Computed, a.Default)
	}
	var resp defaults.StringResponse
	a.Default.DefaultString(context.Background(), defaults.StringRequest{}, &resp)
	if resp.PlanValue.IsNull() || resp.PlanValue.IsUnknown() || resp.PlanValue.ValueString() != "" {
		t.Errorf("description default = %v, want empty string", resp.PlanValue)
	}
}
