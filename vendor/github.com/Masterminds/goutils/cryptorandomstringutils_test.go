package goutils

import (
	"testing"
	"unicode/utf8"
)

func TestCryptoRandomNonAlphaNumeric(t *testing.T) {
	// If asked for a string 0 characters long, CryptoRandomNonAlphaNumeric should provide an empty string.
	if x, _ := CryptoRandomNonAlphaNumeric(0); utf8.RuneCountInString(x) != 0 {
		t.Errorf("String should be 0 characters; string was %v characters", utf8.RuneCountInString(x))
	}

	// Test CryptoRandomNonAlphaNumeric's ability to generate strings 1 through 100 characters in length.
	for i := 1; i < 101; i++ {
		if x, _ := CryptoRandomNonAlphaNumeric(i); utf8.RuneCountInString(x) != i {
			t.Errorf("String should be %v characters; string was %v characters", i, utf8.RuneCountInString(x))
		}
	}
}

func TestCryptoRandomAscii(t *testing.T) {
	// If asked for a string 0 characters long, CryptoRandomAscii should provide an empty string.
	if x, _ := CryptoRandomAscii(0); len(x) != 0 {
		t.Errorf("String should be 0 characters; string was %v characters", len(x))
	}

	// Test CryptoRandomAscii's ability to generate strings 1 through 100 characters in length.
	for i := 1; i < 101; i++ {
		if x, _ := CryptoRandomAscii(i); len(x) != i {
			t.Errorf("String should be %v characters; string was %v characters", i, len(x))
		}
	}
}

func TestCryptoRandomNumeric(t *testing.T) {
	// If asked for a string 0 characters long, CryptoRandomNumeric should provide an empty string.
	if x, _ := CryptoRandomNumeric(0); len(x) != 0 {
		t.Errorf("String should be 0 characters; string was %v characters", len(x))
	}

	// Test CryptoRandomNumeric's ability to generate strings 1 through 100 characters in length.
	for i := 1; i < 101; i++ {
		if x, _ := CryptoRandomNumeric(i); len(x) != i {
			t.Errorf("String should be %v characters; string was %v characters", i, len(x))
		}
	}
}

func TestCryptoRandomAlphabetic(t *testing.T) {
	// If asked for a string 0 characters long, CryptoRandomAlphabetic should provide an empty string.
	if x, _ := CryptoRandomAlphabetic(0); len(x) != 0 {
		t.Errorf("String should be 0 characters; string was %v characters", len(x))
	}

	// Test CryptoRandomAlphabetic's ability to generate strings 1 through 100 characters in length.
	for i := 1; i < 101; i++ {
		if x, _ := CryptoRandomAlphabetic(i); len(x) != i {
			t.Errorf("String should be %v characters; string was %v characters", i, len(x))
		}
	}
}

func TestCryptoRandomAlphaNumeric(t *testing.T) {
	// If asked for a string 0 characters long, CryptoRandomAlphaNumeric should provide an empty string.
	if x, _ := CryptoRandomAlphaNumeric(0); len(x) != 0 {
		t.Errorf("String should be 0 characters; string was %v characters", len(x))
	}

	// Test CryptoRandomAlphaNumeric's ability to generate strings 1 through 100 characters in length.
	for i := 1; i < 101; i++ {
		if x, _ := CryptoRandomAlphaNumeric(i); len(x) != i {
			t.Errorf("String should be %v characters; string was %v characters", i, len(x))
		}
	}
}
