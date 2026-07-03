package ssh

import "testing"

func TestParseMemInfo(t *testing.T) {
	output := `MemTotal:       16384256 kB
MemFree:         1024000 kB
MemAvailable:    8192000 kB
Buffers:          256000 kB
Cached:          1024000 kB`

	total, available, err := parseMemInfo(output)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if total != 16384256*1024 {
		t.Fatalf("unexpected total bytes: %d", total)
	}
	if available != 8192000*1024 {
		t.Fatalf("unexpected available bytes: %d", available)
	}
}

func TestParseMemInfo_FailsWithoutRequiredFields(t *testing.T) {
	_, _, err := parseMemInfo("MemTotal: 123 kB")
	if err == nil {
		t.Fatal("expected error when MemAvailable is missing")
	}
}
