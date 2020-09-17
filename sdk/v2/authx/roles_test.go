package authx

import "testing"

func TestRoleAssignmentMarshalJSON(t *testing.T) {
	requireAPIVersionAndType(t, RoleAssignment{}, "RoleAssignment")
}
