package testing

import (
	"encoding/json"
	"testing"

	"github.com/brigadecore/brigade/sdk/v2/meta"
	"github.com/stretchr/testify/require"
)

func RequireAPIVersionAndType(
	t *testing.T,
	obj interface{},
	expectedType string,
) {
	objJSON, err := json.Marshal(obj)
	require.NoError(t, err)
	objMap := map[string]interface{}{}
	err = json.Unmarshal(objJSON, &objMap)
	require.NoError(t, err)
	require.Equal(t, meta.APIVersion, objMap["apiVersion"])
	require.Equal(t, expectedType, objMap["kind"])
}
