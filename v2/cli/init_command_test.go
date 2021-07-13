package main

import (
	"io/ioutil"
	"path"
	"strings"
	"testing"

	"github.com/brigadecore/brigade/v2/internal/file"
	"github.com/stretchr/testify/require"
)

func TestFileExtensionForLanguage(t *testing.T) {
	testCases := []struct {
		name     string
		expected string
	}{
		{name: "typescript", expected: "ts"},
		{name: "TypeScript", expected: "ts"},
		{name: "ts", expected: "ts"},
		{name: "javascript", expected: "js"},
		{name: "js", expected: "js"},
		{name: "Js", expected: "js"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result, err := fileExtensionForLanguage(testCase.name)
			require.NoError(t, err)
			require.Equal(t, result, testCase.expected)
		})
	}
}

func TestAddLinesToFile(t *testing.T) {
	tempDir := t.TempDir()
	editFilePath := path.Join(tempDir, "editFile")
	err := addLinesToFile(
		editFilePath,
		path.Join(tempDir, "f00bar"),
	)
	require.NoError(t, err)

	// Verify editFile is created
	fileExists, err := file.Exists(editFilePath)
	require.True(t, fileExists)
	require.NoError(t, err)

	// Verify editFile contains f00bar
	verifyFileContents(t, editFilePath, "f00bar")

	// Clear editFile, test functionality of writing to existing editFile
	err = ioutil.WriteFile(editFilePath, []byte(``), 0644)
	require.NoError(t, err)
	err = addLinesToFile(editFilePath, "f00bar")
	require.NoError(t, err)
	verifyFileContents(t, editFilePath, "f00bar")
}

func TestValidateGitCloneURL(t *testing.T) {
	testCases := []struct {
		name     string
		url      string
		expected bool
	}{
		{
			name:     "valid git clone url",
			url:      "https://github.com/brigadecore/brigade.git",
			expected: true,
		},
		{
			name:     "does not start with https://, http://, or git@",
			url:      "github.com/brigadecore/brigade.git",
			expected: false,
		},
		{
			name:     "does not end in .git",
			url:      "https://github.com/brigadecore/brigade",
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
			err := validateGitCloneURL(testCase.url)
			require.Equal(t, testCase.expected, err == nil)
		})
	}
}

func verifyFileContents(t *testing.T, filePath string, content string) {
	b, err := ioutil.ReadFile(filePath)
	require.NoError(t, err)
	s := string(b)
	require.True(t, strings.Contains(s, content))
}
