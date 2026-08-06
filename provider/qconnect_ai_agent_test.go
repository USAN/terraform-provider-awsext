package provider

import (
	"testing"
)

func TestUnmarshalOrchestrationConfig_EmptyVariants(t *testing.T) {
	// JSON with whitespace-padded null/empty values that the old string-literal
	// comparison would fail to recognise as empty.
	toolTemplate := func(overrideInputValues, inputSchema string) string {
		return `{"orchestrationAIPromptId":"p1","toolConfigurations":[{"toolName":"t1","toolType":"FUNCTION","instruction":null,"overrideInputValues":` +
			overrideInputValues + `,"inputSchema":` + inputSchema + `}]}`
	}
	inputs := []string{
		toolTemplate("null", "null"),
		toolTemplate("[]", "null"),
		toolTemplate("[ ]", "null"),
		toolTemplate(" null ", " null "),
	}
	for _, in := range inputs {
		_, err := unmarshalOrchestrationConfig(in)
		if err != nil {
			t.Errorf("unmarshalOrchestrationConfig(%q) returned unexpected error: %v", in, err)
		}
	}
}

func TestUnmarshalOrchestrationConfig_WithValues(t *testing.T) {
	in := `{
		"orchestrationAIPromptId": "prompt-abc",
		"toolConfigurations": [{
			"toolName": "myTool",
			"toolType": "FUNCTION",
			"instruction": null,
			"overrideInputValues": [{"jsonPath":"$.k","value":{"constant":{"type":"STRING","value":"v"}}}],
			"inputSchema": {"type":"object"}
		}]
	}`
	cfg, err := unmarshalOrchestrationConfig(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.ToolConfigurations) == 0 {
		t.Fatal("expected ToolConfigurations to be populated")
	}
	if len(cfg.ToolConfigurations[0].OverrideInputValues) == 0 {
		t.Error("expected OverrideInputValues to be populated")
	}
	if cfg.ToolConfigurations[0].InputSchema == nil {
		t.Error("expected InputSchema to be populated")
	}
}

func TestIsNonEmptyJSONArray(t *testing.T) {
	cases := []struct {
		input    []byte
		expected bool
	}{
		{nil, false},
		{[]byte(""), false},
		{[]byte("null"), false},
		{[]byte("[ ]"), false},    // whitespace-only array
		{[]byte("[]"), false},
		{[]byte("{}"), false},     // object, not array
		{[]byte("garbage"), false},
		{[]byte(`[{"name":"k"}]`), true},
		{[]byte(`[ {"name":"k"} ]`), true}, // whitespace around elements
	}
	for _, c := range cases {
		got := isNonEmptyJSONArray(c.input)
		if got != c.expected {
			t.Errorf("isNonEmptyJSONArray(%q) = %v, want %v", c.input, got, c.expected)
		}
	}
}

func TestIsNonNullJSON(t *testing.T) {
	cases := []struct {
		input    []byte
		expected bool
	}{
		{nil, false},
		{[]byte(""), false},
		{[]byte("null"), false},
		{[]byte(" null "), false},  // whitespace-padded null
		{[]byte("garbage"), false},
		{[]byte("{}"), true},
		{[]byte(`{"type":"object"}`), true},
		{[]byte("[]"), true},
		{[]byte("0"), true},
		{[]byte(`"string"`), true},
	}
	for _, c := range cases {
		got := isNonNullJSON(c.input)
		if got != c.expected {
			t.Errorf("isNonNullJSON(%q) = %v, want %v", c.input, got, c.expected)
		}
	}
}
