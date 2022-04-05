package meta

import (
	"testing"

	metaTesting "github.com/brigadecore/brigade/v2/apiserver/internal/meta/testing" // nolint: lll
	"github.com/stretchr/testify/require"
)

func TestListLen(t *testing.T) {
	list := List[int]{
		Items: []int{2, 3, 5, 1, 4},
	}
	require.Equal(t, int64(len(list.Items)), list.Len())
}

func TestListSort(t *testing.T) {
	list := List[int]{
		Items: []int{2, 3, 5, 1, 4},
	}
	list.Sort(func(lhs, rhs int) int {
		return lhs - rhs
	})
	require.Equal(t, []int{1, 2, 3, 4, 5}, list.Items)
}

func TestListMarshalJSON(t *testing.T) {
	type TestType struct{}
	metaTesting.RequireAPIVersionAndType(t, List[TestType]{}, "TestTypeList")
}
