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
			name: "types do not match",
			roleAssignment: RoleAssignment{
				Role: Role{
					Type: "foo",
					Name: "foo",
				},
				Scope: "foo",
			},
			role: Role{
				Type: "bar",
				Name: "foo",
			},
			scope:   "foo",
			matches: false,
		},
		{
			name: "names do not match",
			roleAssignment: RoleAssignment{
				Role: Role{
					Type: "foo",
					Name: "foo",
				},
				Scope: "foo",
			},
			role: Role{
				Type: "foo",
				Name: "bar",
			},
			scope:   "foo",
			matches: false,
		},
		{
			name: "scopes do not match",
			roleAssignment: RoleAssignment{
				Role: Role{
					Type: "foo",
					Name: "foo",
				},
				Scope: "foo",
			},
			role: Role{
				Type: "foo",
				Name: "foo",
			},
			scope:   "bar",
			matches: false,
		},
		{
			name: "scopes are an exact match",
			roleAssignment: RoleAssignment{
				Role: Role{
					Type: "foo",
					Name: "foo",
				},
				Scope: "foo",
			},
			role: Role{
				Type: "foo",
				Name: "foo",
			},
			scope:   "foo",
			matches: true,
		},
		{
			name: "a global scope matches b scope",
			roleAssignment: RoleAssignment{
				Role: Role{
					Type: "foo",
					Name: "foo",
				},
				Scope: RoleScopeGlobal,
			},
			role: Role{
				Type: "foo",
				Name: "foo",
			},
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
