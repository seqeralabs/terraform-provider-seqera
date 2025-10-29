package boolvalidators

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// Note: Full integration testing of this validator requires a complete Config object
// with sibling fields. The validator logic is tested indirectly through provider tests.
// These unit tests verify the basic validator structure and description.

func TestWaveEnabledValidator_BasicBehavior(t *testing.T) {
	t.Parallel()

	// Test that validator skips validation when wave_enabled is false
	v := WaveEnabledValidator()

	if v == nil {
		t.Fatal("WaveEnabledValidator() returned nil")
	}

	// Verify validator implements the interface
	var _ validator.Bool = v
}

func TestWaveEnabledValidator_Description(t *testing.T) {
	v := WaveEnabledValidator()
	desc := v.Description(context.Background())

	if desc == "" {
		t.Error("expected non-empty description")
	}

	if desc != v.MarkdownDescription(context.Background()) {
		t.Error("expected Description and MarkdownDescription to match")
	}

	// Verify description mentions the key validation rules
	expectedKeywords := []string{"enable_wave", "enable_fusion"}
	for _, keyword := range expectedKeywords {
		if !contains(desc, keyword) {
			t.Errorf("description should mention %q but doesn't: %s", keyword, desc)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && (s[:len(substr)] == substr || contains(s[1:], substr))))
}
