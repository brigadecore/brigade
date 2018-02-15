package kube

import "testing"

func TestSecretValues(t *testing.T) {
	sv := SecretValues{
		"magic":  []byte("black"),
		"byteme": nil,
	}

	if got := sv.String("magic"); got != "black" {
		t.Errorf("%v", got)
	}
	if got := sv.String("byteme"); got != "" {
		t.Errorf("%v", got)
	}
	if got := sv.Bytes("nil"); got != nil {
		t.Errorf("%v", got)
	}
	if noop := sv.String("noop"); noop != "" {
		t.Errorf("%v", noop)
	}
}
