package sdk

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateProjectID(t *testing.T) {
	testCases := []struct {
		name     string
		id       string
		expected bool
	}{
		{
			name:     "valid project id",
			id:       "abc",
			expected: true,
		},
		{
			name:     "id is too short",
			id:       "aa",
			expected: false,
		},
		{
			name: "id is too long by 1 character",
			id: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" +
				"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			expected: false,
		},
		{
			name:     "id starts with a non-letter character",
			id:       "1aa",
			expected: false,
		},
		{
			name:     "id contains a capital letter",
			id:       "Hello",
			expected: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.id, func(t *testing.T) {
			err := ValidateProjectID(testCase.id)
			require.Equal(t, testCase.expected, err == nil)
		})
	}
}

func TestValidateGitCloneURL(t *testing.T) {
	testCases := []struct {
		name     string
		url      string
		expected bool
	}{
		{
			name:     "valid git clone url",
			url:      "https://github.com/brigadecore/brigade?foo=bat%%20baz",
			expected: true,
		},
		{
			name:     "does not start with https://, http://, or git@",
			url:      "github.com/brigadecore/brigade.git",
			expected: false,
		},
		{
			name:     "not a link",
			url:      "foobar",
			expected: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.url, func(t *testing.T) {
			err := ValidateGitCloneURL(testCase.url)
			require.Equal(t, testCase.expected, err == nil)
		})
	}
}
