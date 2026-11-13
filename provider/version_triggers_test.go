package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// TestVersionResourcesForceReplacementOnTriggers guards the property the
// republish logic depends on: changing `triggers` must plan a replacement.
func TestVersionResourcesForceReplacementOnTriggers(t *testing.T) {
	cases := map[string]resource.Resource{
		"prompt version":    NewQConnectAIPromptVersionResource(),
		"guardrail version": NewQConnectAIGuardrailVersionResource(),
		"agent version":     NewQConnectAIAgentVersionResource(),
	}
	for name, res := range cases {
		t.Run(name, func(t *testing.T) {
			var resp resource.SchemaResponse
			res.Schema(context.Background(), resource.SchemaRequest{}, &resp)

			attr, ok := resp.Schema.Attributes["triggers"].(schema.MapAttribute)
			if !ok {
				t.Fatalf("triggers is missing or not a map attribute")
			}
			if !attr.Optional {
				t.Errorf("triggers must be optional so existing configs keep working")
			}
			if len(attr.PlanModifiers) != 1 {
				t.Errorf("triggers must carry exactly the RequiresReplace plan modifier, got %d", len(attr.PlanModifiers))
			}
		})
	}
}
