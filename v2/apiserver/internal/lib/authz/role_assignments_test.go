package authz

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMatches(t *testing.T) {
	testCases := []struct {
		name           string
		roleAssignment RoleAssignment
		role           Role
		scope          string
		matches        bool
	}{
		{
			name: "names do not match",
			roleAssignment: RoleAssignment{
				Role:  "foo",
				Scope: "foo",
			},
			role:    "bar",
			scope:   "foo",
			matches: false,
		},
		{
			name: "scopes do not match",
			roleAssignment: RoleAssignment{
				Role:  "foo",
				Scope: "foo",
			},
			role:    "foo",
			scope:   "bar",
			matches: false,
		},
		{
			name: "scopes are an exact match",
			roleAssignment: RoleAssignment{
				Role:  "foo",
				Scope: "foo",
			},
			role:    "foo",
			scope:   "foo",
			matches: true,
		},
		{
			name: "a global scope matches b scope",
			roleAssignment: RoleAssignment{
				Role:  "foo",
				Scope: RoleScopeGlobal,
			},
			role:    "foo",
			scope:   "foo",
			matches: true,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			require.Equal(
				t,
				testCase.matches,
				testCase.roleAssignment.Matches(testCase.role, testCase.scope),
			)
		})
	}
}
