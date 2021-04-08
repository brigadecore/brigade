package authz

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMatches(t *testing.T) {
	testCases := []struct {
		name    string
		a       Role
		b       Role
		matches bool
	}{
		{
			name: "names do not match",
			a: Role{
				Name:  "foo",
				Scope: "foo",
			},
			b: Role{
				Name:  "bar",
				Scope: "foo",
			},
			matches: false,
		},
		{
			name: "scopes do not match",
			a: Role{
				Name:  "foo",
				Scope: "foo",
			},
			b: Role{
				Name:  "foo",
				Scope: "bar",
			},
			matches: false,
		},
		{
			name: "scopes are an exact match",
			a: Role{
				Name:  "foo",
				Scope: "foo",
			},
			b: Role{
				Name:  "foo",
				Scope: "foo",
			},
			matches: true,
		},
		{
			name: "a global scope matches b scope",
			a: Role{
				Name:  "foo",
				Scope: RoleScopeGlobal,
			},
			b: Role{
				Name:  "foo",
				Scope: "foo",
			},
			matches: true,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			require.Equal(t, testCase.matches, testCase.a.Matches(testCase.b))
		})
	}
}
