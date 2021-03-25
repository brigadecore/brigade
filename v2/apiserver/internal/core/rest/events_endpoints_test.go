package rest

import (
	"net/url"
	"testing"

	"github.com/brigadecore/brigade/v2/apiserver/internal/core"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/stretchr/testify/require"
)

func TestEventsSelectorFromURLQuery(t *testing.T) {
	testCases := []struct {
		name        string
		queryParams url.Values
		assertions  func(core.EventsSelector, *meta.ErrBadRequest)
	}{
		{
			name: "nil query params",
			assertions: func(selector core.EventsSelector, err *meta.ErrBadRequest) {
				require.Nil(t, err)
				require.Equal(t, core.EventsSelector{}, selector)
			},
		},
		{
			name: "invalid source state",
			queryParams: url.Values{
				"sourceState": []string{"key-value"},
			},
			assertions: func(selector core.EventsSelector, err *meta.ErrBadRequest) {
				require.Contains(t, err.Error(), `Invalid value "key-value"`)
			},
		},
		{
			name: "success",
			queryParams: url.Values{
				"projectID":    []string{"blue-book"},
				"source":       []string{"brigade.sh/cli"},
				"sourceState":  []string{"foo=bar,bat=baz"},
				"type":         []string{"exec"},
				"workerPhases": []string{"PENDING,STARTING"},
			},
			assertions: func(selector core.EventsSelector, err *meta.ErrBadRequest) {
				require.Nil(t, err)
				require.Equal(
					t,
					core.EventsSelector{
						ProjectID: "blue-book",
						Source:    "brigade.sh/cli",
						SourceState: map[string]string{
							"foo": "bar",
							"bat": "baz",
						},
						Type: "exec",
						WorkerPhases: []core.WorkerPhase{
							core.WorkerPhasePending,
							core.WorkerPhaseStarting,
						},
					},
					selector,
				)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.assertions(
				eventsSelectorFromURLQuery(testCase.queryParams),
			)
		})
	}
}
