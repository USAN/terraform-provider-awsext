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
