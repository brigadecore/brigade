package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateLabel(t *testing.T) {
	generated := GenerateLabel(
		"my-suuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuper-sweet-project",
	)
	require.Equal(t, 63, len(generated))
	require.Equal(
		t,
		"my-suuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuper...roject",
		generated,
	)
}
