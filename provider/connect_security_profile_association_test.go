package provider

import (
	"errors"
	"testing"

	conntypes "github.com/aws/aws-sdk-go-v2/service/connect/types"
)

func TestSecurityProfileAssociationScopes(t *testing.T) {
	base := "arn:aws:wisdom:us-east-1:111111111111:ai-agent/assistant-id/agent-id"
	got := securityProfileAssociationScopes(base)
	want := []string{
		base,
		base + ":$LATEST",
		base + ":$SAVED",
	}
	if len(got) != len(want) {
		t.Fatalf("got %d scopes, want %d: %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("scope[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestConnectSecurityProfileAssociationID(t *testing.T) {
	got := connectSecurityProfileAssociationID("inst-1", "sp-1", "arn:aws:wisdom:us-east-1:111111111111:ai-agent/a/b")
	want := "inst-1/sp-1/arn:aws:wisdom:us-east-1:111111111111:ai-agent/a/b"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestIsResourceConflict(t *testing.T) {
	var confErr *conntypes.ResourceConflictException = &conntypes.ResourceConflictException{}
	if !isResourceConflict(confErr) {
		t.Error("expected isResourceConflict to be true for *ResourceConflictException")
	}
	if isResourceConflict(errors.New("some other error")) {
		t.Error("expected isResourceConflict to be false for an unrelated error")
	}
	if isResourceConflict(nil) {
		t.Error("expected isResourceConflict to be false for nil")
	}
}

func TestIsResourceNotFound(t *testing.T) {
	var nfErr *conntypes.ResourceNotFoundException = &conntypes.ResourceNotFoundException{}
	if !isResourceNotFound(nfErr) {
		t.Error("expected isResourceNotFound to be true for *ResourceNotFoundException")
	}
	if isResourceNotFound(errors.New("some other error")) {
		t.Error("expected isResourceNotFound to be false for an unrelated error")
	}
	if isResourceNotFound(nil) {
		t.Error("expected isResourceNotFound to be false for nil")
	}
}

func TestParseAIAgentArn(t *testing.T) {
	const base = "arn:aws:wisdom:us-east-1:111111111111:ai-agent/assistant-1/agent-1"
	tests := []struct {
		name          string
		arn           string
		wantAssistant string
		wantAgent     string
		wantOK        bool
	}{
		{"base arn", base, "assistant-1", "agent-1", true},
		{"version qualifier", base + ":6", "assistant-1", "agent-1", true},
		{"latest qualifier", base + ":$LATEST", "assistant-1", "agent-1", true},
		{"not an agent arn", "arn:aws:wisdom:us-east-1:111111111111:assistant/assistant-1", "", "", false},
		{"missing agent id", "arn:aws:wisdom:us-east-1:111111111111:ai-agent/assistant-1", "", "", false},
		{"empty", "", "", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assistant, agent, ok := parseAIAgentArn(tt.arn)
			if assistant != tt.wantAssistant || agent != tt.wantAgent || ok != tt.wantOK {
				t.Errorf("parseAIAgentArn(%q) = (%q, %q, %v), want (%q, %q, %v)",
					tt.arn, assistant, agent, ok, tt.wantAssistant, tt.wantAgent, tt.wantOK)
			}
		})
	}
}
