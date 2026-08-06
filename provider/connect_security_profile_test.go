package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestPreserveOrSetStringList(t *testing.T) {
	ctx := context.Background()

	t.Run("preserves null when items empty and current is null", func(t *testing.T) {
		current := types.ListNull(types.StringType)
		result, diags := preserveOrSetStringList(ctx, []string{}, current)
		if diags.HasError() {
			t.Fatalf("unexpected diags: %v", diags)
		}
		if !result.IsNull() {
			t.Error("expected null list to be preserved")
		}
	})

	t.Run("sets empty list when items empty and current is not null", func(t *testing.T) {
		current := types.ListValueMust(types.StringType, nil)
		result, diags := preserveOrSetStringList(ctx, []string{}, current)
		if diags.HasError() {
			t.Fatalf("unexpected diags: %v", diags)
		}
		if result.IsNull() {
			t.Error("expected empty (non-null) list")
		}
		if len(result.Elements()) != 0 {
			t.Errorf("expected 0 elements, got %d", len(result.Elements()))
		}
	})

	t.Run("sets populated list when items present", func(t *testing.T) {
		current := types.ListNull(types.StringType)
		result, diags := preserveOrSetStringList(ctx, []string{"a", "b"}, current)
		if diags.HasError() {
			t.Fatalf("unexpected diags: %v", diags)
		}
		if result.IsNull() {
			t.Error("expected non-null list")
		}
		if len(result.Elements()) != 2 {
			t.Errorf("expected 2 elements, got %d", len(result.Elements()))
		}
	})
}
