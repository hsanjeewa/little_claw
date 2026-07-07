package ssh

import "testing"

func TestParseNetworkDevOutput(t *testing.T) {
	output := `Inter-|   Receive                                                |  Transmit
 face |bytes    packets errs drop fifo frame compressed multicast|bytes    packets errs drop fifo colls carrier compressed
  eth0: 123456789  1234    0    0    0     0          0         0 987654321  5678    0    0    0     0       0         0
    lo: 99999999   999    0    0    0     0          0         0 88888888   888    0    0    0     0       0         0`
	rx, tx, err := parseNetworkDevOutput(output)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rx != 123456789 {
		t.Fatalf("unexpected rx bytes: %d, want %d", rx, 123456789)
	}
	if tx != 987654321 {
		t.Fatalf("unexpected tx bytes: %d, want %d", tx, 987654321)
	}
}

func TestParseNetworkDevOutput_Empty(t *testing.T) {
	_, _, err := parseNetworkDevOutput("")
	if err == nil {
		t.Fatal("expected error for empty output")
	}
}

func TestParseNetworkDevOutput_OnlyLoopback(t *testing.T) {
	output := `Inter-|   Receive                                                |  Transmit
 face |bytes    packets errs drop fifo frame compressed multicast|bytes    packets errs drop fifo colls carrier compressed
    lo: 99999999   999    0    0    0     0          0         0 88888888   888    0    0    0     0       0         0`
	_, _, err := parseNetworkDevOutput(output)
	if err == nil {
		t.Fatal("expected error when only loopback exists")
	}
}

func TestParseNetworkDevOutput_MalformedLine(t *testing.T) {
	output := `Inter-|   Receive                                                |  Transmit
 face |bytes    packets errs drop fifo frame compressed multicast|bytes    packets errs drop fifo colls carrier compressed
  eth0: notanumber  1234    0    0    0     0          0         0 987654321  5678    0    0    0     0       0         0`
	_, _, err := parseNetworkDevOutput(output)
	if err == nil {
		t.Fatal("expected error for malformed rx value")
	}
}
