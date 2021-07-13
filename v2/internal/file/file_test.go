package file

import (
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExists(t *testing.T) {
	file := "file_test.go"
	exists, err := Exists(file)
	require.NoError(t, err)
	require.True(t, exists)

	file = "bogus.go"
	exists, err = Exists(file)
	require.NoError(t, err)
	require.False(t, exists)
}

func TestEnsureDirectory(t *testing.T) {
	tempDir := t.TempDir()
	tempPath := path.Join(tempDir, ".brigade")

	// Test for expected behavior when .brigade does not exist
	created, err := EnsureDirectory(tempPath)
	require.NoError(t, err)
	require.DirExists(t, tempPath)
	require.False(t, created)

	// Test for expected behavior when .brigade exists
	created, err = EnsureDirectory(tempPath)
	require.NoError(t, err)
	require.DirExists(t, tempPath)
	require.True(t, created)
}
