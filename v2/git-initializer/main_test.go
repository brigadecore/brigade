package main

import (
	"testing"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/stretchr/testify/require"
)

func Test_isFullReference(t *testing.T) {
	t.Run("full branch reference", func(t *testing.T) {
		ref := plumbing.ReferenceName("refs/heads/foo")
		require.True(t, isFullReference(ref))
	})

	t.Run("full tag reference", func(t *testing.T) {
		ref := plumbing.ReferenceName("refs/tags/foo")
		require.True(t, isFullReference(ref))
	})

	t.Run("full remote reference", func(t *testing.T) {
		ref := plumbing.ReferenceName("refs/remotes/foo")
		require.True(t, isFullReference(ref))
	})

	t.Run("full note reference", func(t *testing.T) {
		ref := plumbing.ReferenceName("refs/notes/foo")
		require.True(t, isFullReference(ref))
	})

	t.Run("short reference", func(t *testing.T) {
		ref := plumbing.ReferenceName("foo")
		require.False(t, isFullReference(ref))
	})

	t.Run("tasty reference", func(t *testing.T) {
		ref := plumbing.ReferenceName("refs/tacos/carnitas")
		require.False(t, isFullReference(ref))
	})
}
