package core

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMatches(t *testing.T) {
	testCases := []struct {
		name    string
		a       ProjectRole
		b       ProjectRole
		matches bool
	}{
		{
			name: "names do not match",
			a: ProjectRole{
				Name:      "foo",
				ProjectID: "foo",
			},
			b: ProjectRole{
				Name:      "bar",
				ProjectID: "foo",
			},
			matches: false,
		},
		{
			name: "projectIDs do not match",
			a: ProjectRole{
				Name:      "foo",
				ProjectID: "foo",
			},
			b: ProjectRole{
				Name:      "foo",
				ProjectID: "bar",
			},
			matches: false,
		},
		{
			name: "projectIDs are an exact match",
			a: ProjectRole{
				Name:      "foo",
				ProjectID: "foo",
			},
			b: ProjectRole{
				Name:      "foo",
				ProjectID: "foo",
			},
			matches: true,
		},
		{
			name: "a global projectID matches b projectID",
			a: ProjectRole{
				Name:      "foo",
				ProjectID: ProjectIDGlobal,
			},
			b: ProjectRole{
				Name:      "foo",
				ProjectID: "foo",
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
