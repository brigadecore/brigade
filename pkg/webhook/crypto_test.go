package webhook

import (
	"testing"
)

func TestSHA1HMAC(t *testing.T) {
	salt := []byte("This is the way the world ends.")
	message := []byte("Not with a bang, but a whimper.\n")
	expect := "sha1=0ca6713b350828f53c6dcced9232aeace3e60708"
	if got := SHA1HMAC(salt, message); got != expect {
		t.Fatalf("Expected \n\t%q, got\n\t%q", expect, got)
	}
}
