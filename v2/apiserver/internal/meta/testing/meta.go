package testing

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

const apiVersion = "brigade.sh/v2-alpha.4"

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
	require.Equal(t, apiVersion, objMap["apiVersion"])
	require.Equal(t, expectedType, objMap["kind"])
}
