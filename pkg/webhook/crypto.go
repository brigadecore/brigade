package webhook

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
)

// shortSha returns a 32-char SHA256 digest as a string.
func shortSha(input string) string {
	sum := sha256.Sum256([]byte(input))
	return fmt.Sprintf("%x", sum)[0:54]
}

// Compute the GitHub SHA1 HMAC.
func sha1HMAC(salt, message []byte) string {
	// GitHub creates a SHA1 HMAC, where the key is the GitHub secret and the
	// message is the JSON body.
	digest := hmac.New(sha1.New, salt)
	digest.Write(message)
	sum := digest.Sum(nil)
	return fmt.Sprintf("sha1=%x", sum)
}
