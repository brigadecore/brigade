package authz

import (
	"testing"

	metaTesting "github.com/brigadecore/brigade/sdk/v3/meta/testing"
)

func TestRoleAssignmentMarshalJSON(t *testing.T) {
	metaTesting.RequireAPIVersionAndType(t, RoleAssignment{}, RoleAssignmentKind)
}
