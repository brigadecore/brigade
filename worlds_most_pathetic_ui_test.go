package main

import (
	"testing"
)

func TestSHAish(t *testing.T) {
	tests := map[string]bool{
		"66368814580485ac6422a38e0a857bd9a3b0a64f":    true,
		"66368814580485ac6422a38e0a857bd9a3b0a64":     false,
		"66368814580485ac6422a38e0a857bd9a3b0a64ffff": false,
		"XXXXXXXXA80485ac6422a38e0a857bd9a3b0a64f":    false,
		"master": false,
	}

	for sha, expect := range tests {
		if SHAish(sha) != expect {
			t.Errorf("expected %s to be %v", sha, expect)
		}
	}
}
