package provider

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appintegrations"
	appintegrationstypes "github.com/aws/aws-sdk-go-v2/service/appintegrations/types"
	smithy "github.com/aws/smithy-go"
)

type fakeAssociationLister struct {
	// responses is consumed in order, one per ListApplicationAssociations call;
	// the last entry repeats once exhausted.
	responses []fakeAssociationResponse
	calls     int
}

type fakeAssociationResponse struct {
	associations []appintegrationstypes.ApplicationAssociationSummary
	err          error
}

func (f *fakeAssociationLister) ListApplicationAssociations(ctx context.Context, params *appintegrations.ListApplicationAssociationsInput, optFns ...func(*appintegrations.Options)) (*appintegrations.ListApplicationAssociationsOutput, error) {
	i := f.calls
	if i >= len(f.responses) {
		i = len(f.responses) - 1
	}
	f.calls++
	r := f.responses[i]
	if r.err != nil {
		return nil, r.err
	}
	return &appintegrations.ListApplicationAssociationsOutput{ApplicationAssociations: r.associations}, nil
}

// resourceNotFoundError satisfies errors.As(err, *appintegrationstypes.ResourceNotFoundException)
// via smithy's APIError wrapping, matching what the SDK actually returns.
func resourceNotFoundError() error {
	return &appintegrationstypes.ResourceNotFoundException{Message: aws.String("not found")}
}

var _ smithy.APIError = &appintegrationstypes.ResourceNotFoundException{}

func TestWaitForNoApplicationAssociations_EmptyImmediately(t *testing.T) {
	lister := &fakeAssociationLister{responses: []fakeAssociationResponse{{}}}

	err := waitForNoApplicationAssociations(context.Background(), lister, "app-1", time.Millisecond, 50*time.Millisecond)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if lister.calls != 1 {
		t.Fatalf("expected exactly 1 call when already empty, got %d", lister.calls)
	}
}

func TestWaitForNoApplicationAssociations_ClearsAfterRetry(t *testing.T) {
	lister := &fakeAssociationLister{responses: []fakeAssociationResponse{
		{associations: []appintegrationstypes.ApplicationAssociationSummary{{ApplicationArn: aws.String("app-1")}}},
		{associations: []appintegrationstypes.ApplicationAssociationSummary{{ApplicationArn: aws.String("app-1")}}},
		{}, // clears on 3rd poll
	}}

	err := waitForNoApplicationAssociations(context.Background(), lister, "app-1", 2*time.Millisecond, 200*time.Millisecond)
	if err != nil {
		t.Fatalf("expected nil error once associations clear, got %v", err)
	}
	if lister.calls != 3 {
		t.Fatalf("expected 3 calls, got %d", lister.calls)
	}
}

func TestWaitForNoApplicationAssociations_TimesOutWhenStuck(t *testing.T) {
	lister := &fakeAssociationLister{responses: []fakeAssociationResponse{
		{associations: []appintegrationstypes.ApplicationAssociationSummary{{ApplicationArn: aws.String("app-1")}}},
	}}

	err := waitForNoApplicationAssociations(context.Background(), lister, "app-1", 5*time.Millisecond, 20*time.Millisecond)
	if err == nil {
		t.Fatal("expected timeout error when associations never clear, got nil")
	}
}

func TestWaitForNoApplicationAssociations_ResourceNotFoundIsSuccess(t *testing.T) {
	lister := &fakeAssociationLister{responses: []fakeAssociationResponse{
		{err: resourceNotFoundError()},
	}}

	err := waitForNoApplicationAssociations(context.Background(), lister, "app-1", time.Millisecond, 50*time.Millisecond)
	if err != nil {
		t.Fatalf("expected ResourceNotFoundException to be treated as success, got %v", err)
	}
}

func TestWaitForNoApplicationAssociations_OtherErrorPropagates(t *testing.T) {
	boom := errors.New("boom")
	lister := &fakeAssociationLister{responses: []fakeAssociationResponse{{err: boom}}}

	err := waitForNoApplicationAssociations(context.Background(), lister, "app-1", time.Millisecond, 50*time.Millisecond)
	if err == nil || !errors.Is(err, boom) {
		t.Fatalf("expected wrapped boom error, got %v", err)
	}
}

func TestWaitForNoApplicationAssociations_NoOpWithoutID(t *testing.T) {
	lister := &fakeAssociationLister{}

	err := waitForNoApplicationAssociations(context.Background(), lister, "", time.Millisecond, 50*time.Millisecond)
	if err != nil {
		t.Fatalf("expected nil error with empty applicationID, got %v", err)
	}
	if lister.calls != 0 {
		t.Fatalf("expected no API calls with empty applicationID, got %d", lister.calls)
	}
}

func TestWaitForNoApplicationAssociations_ContextCancelled(t *testing.T) {
	lister := &fakeAssociationLister{responses: []fakeAssociationResponse{
		{associations: []appintegrationstypes.ApplicationAssociationSummary{{ApplicationArn: aws.String("app-1")}}},
	}}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := waitForNoApplicationAssociations(ctx, lister, "app-1", 10*time.Millisecond, time.Second)
	if err == nil {
		t.Fatal("expected context cancellation error, got nil")
	}
}
