package authz

import (
	"testing"

	metaTesting "github.com/brigadecore/brigade/sdk/v2/meta/testing"
)

func TestRoleAssignmentMarshalJSON(t *testing.T) {
	metaTesting.RequireAPIVersionAndType(t, RoleAssignment{}, RoleAssignmentKind)
}
