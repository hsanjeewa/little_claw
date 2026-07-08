package agent

import (
	"context"
	"testing"
)

type MockDiscoverer struct {
	Called bool
}

func (m *MockDiscoverer) Discover(ctx context.Context, alias string) (HostProfile, error) {
	m.Called = true
	return NewHostProfile(alias), nil
}

func TestCapabilityDiscovery(t *testing.T) {
	mock := &MockDiscoverer{}
	ctx := context.Background()

	profile, err := mock.Discover(ctx, "web-prod-01")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !mock.Called {
		t.Error("Expected Discover to be called")
	}

	if profile.HostAlias != "web-prod-01" {
		t.Errorf("Expected web-prod-01, got %s", profile.HostAlias)
	}
}
