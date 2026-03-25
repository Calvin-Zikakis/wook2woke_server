package server

import (
	"encoding/hex"
	"testing"
)

func TestRandomHex_Length(t *testing.T) {
	for _, n := range []int{1, 8, 16, 32} {
		got := randomHex(n)
		// Each byte encodes to 2 hex chars
		if len(got) != n*2 {
			t.Fatalf("randomHex(%d): expected length %d, got %d", n, n*2, len(got))
		}
	}
}

func TestRandomHex_ValidHex(t *testing.T) {
	got := randomHex(16)
	if _, err := hex.DecodeString(got); err != nil {
		t.Fatalf("randomHex returned invalid hex string %q: %v", got, err)
	}
}

func TestRandomHex_Unique(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		v := randomHex(16)
		if seen[v] {
			t.Fatalf("randomHex returned duplicate value %q", v)
		}
		seen[v] = true
	}
}
