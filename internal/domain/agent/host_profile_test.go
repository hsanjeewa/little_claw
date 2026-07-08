package agent

import (
	"testing"
)

func TestHostProfile_Creation(t *testing.T) {
	profile := NewHostProfile("web-prod-01")
	if profile.HostAlias != "web-prod-01" {
		t.Errorf("Expected alias web-prod-01, got %s", profile.HostAlias)
	}
}
