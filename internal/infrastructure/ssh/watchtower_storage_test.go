package ssh

import "testing"

func TestParseDFOutput(t *testing.T) {
	output := `10240000 5120000`
	total, used, err := parseDFOutput(output)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 10240000*1024 {
		t.Fatalf("unexpected total bytes: %d, want %d", total, 10240000*1024)
	}
	if used != 5120000*1024 {
		t.Fatalf("unexpected used bytes: %d, want %d", used, 5120000*1024)
	}
}

func TestParseDFOutput_FailsOnEmpty(t *testing.T) {
	_, _, err := parseDFOutput("")
	if err == nil {
		t.Fatal("expected error for empty output")
	}
}

func TestParseDFOutput_FailsOnMissingFields(t *testing.T) {
	_, _, err := parseDFOutput("total used")
	if err == nil {
		t.Fatal("expected error for missing fields")
	}
}
