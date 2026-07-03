package ssh

import "testing"

func TestParseCPUPercent(t *testing.T) {
	usage, err := parseCPUPercent("37.5\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if usage != 37.5 {
		t.Fatalf("unexpected cpu usage: %v", usage)
	}
}

func TestParseCPUPercent_ClampsToRange(t *testing.T) {
	usage, err := parseCPUPercent("120")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if usage != 100 {
		t.Fatalf("expected clamped cpu usage 100, got %v", usage)
	}
}
