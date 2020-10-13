package crypto

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/brigadecore/brigade/v2/internal/rand"
	uuid "github.com/satori/go.uuid"
)

var seededRand = rand.NewSeeded()

// Hash returns a secure hash of the provided input. If a salt is provided, the
// input is prepended with the salt prior to hashing.
func Hash(salt, input string) string {
	if salt != "" {
		input = fmt.Sprintf("%s:%s", salt, input)
	}
	sum := sha256.Sum256([]byte(input))
	return fmt.Sprintf("%x", sum)
}

// NewToken returns a new, unique, hard-to-guess token consisting of
// alphanumeric characters only. The token's length is the greater of 64 or the
// length specified. i.e. This function produces tokens with a minimum length of
// 64 characters. The generated tokens incorporate a UUID (v4) to ensure
// uniqueness. This minimum token length of 64 was selected so that any given
// token consists of at least as many random, hard-to-guess characters is it
// does more deterministic characters.
func NewToken(tokenLength int) string {
	if tokenLength < 64 {
		tokenLength = 64
	}
	const (
		tokenChars = "abcdefghijklmnopqrstuvwxyz" +
			"ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
			"0123456789"
	)
	const tokenCharsLen = len(tokenChars)

	// Incorporating a UUID into the token guarantees uniqueness.
	tokenPrefix := strings.ReplaceAll(uuid.NewV4().String(), "-", "")

	remainingTokenLength := tokenLength - len(tokenPrefix)

	// Round out the token with random characters...
	b := make([]byte, remainingTokenLength)
	for i := 0; i < remainingTokenLength; i++ {
		b[i] = tokenChars[seededRand.Intn(tokenCharsLen)]
	}

	return fmt.Sprintf("%s%s", tokenPrefix, string(b))
}
