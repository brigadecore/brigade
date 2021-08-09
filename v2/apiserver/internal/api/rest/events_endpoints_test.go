package rest

import (
	"net/url"
	"testing"

	"github.com/brigadecore/brigade/v2/apiserver/internal/api"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/stretchr/testify/require"
)

func TestEventsSelectorFromURLQuery(t *testing.T) {
	testCases := []struct {
		name        string
		queryParams url.Values
		assertions  func(api.EventsSelector, *meta.ErrBadRequest)
	}{
		{
			name: "nil query params",
			assertions: func(selector api.EventsSelector, err *meta.ErrBadRequest) {
				require.Nil(t, err)
				require.Equal(t, api.EventsSelector{}, selector)
			},
		},
		{
			name: "invalid qualifiers",
			queryParams: url.Values{
				"qualifiers": []string{"key-value"},
			},
			assertions: func(selector api.EventsSelector, err *meta.ErrBadRequest) {
				require.Contains(t, err.Error(), `Invalid value "key-value"`)
			},
		},
		{
			name: "invalid labels",
			queryParams: url.Values{
				"labels": []string{"key-value"},
			},
			assertions: func(selector api.EventsSelector, err *meta.ErrBadRequest) {
				require.Contains(t, err.Error(), `Invalid value "key-value"`)
			},
		},
		{
			name: "invalid source state",
			queryParams: url.Values{
				"sourceState": []string{"key-value"},
			},
			assertions: func(selector api.EventsSelector, err *meta.ErrBadRequest) {
				require.Contains(t, err.Error(), `Invalid value "key-value"`)
			},
		},
		{
			name: "success",
			queryParams: url.Values{
				"projectID":    []string{"blue-book"},
				"source":       []string{"brigade.sh/cli"},
				"qualifiers":   []string{"foo=bar,bat=baz"},
				"labels":       []string{"abc=easy-as,123=do-rei-mei"},
				"sourceState":  []string{"baby=you,and=me-girl"},
				"type":         []string{"exec"},
				"workerPhases": []string{"PENDING,STARTING"},
			},
			assertions: func(selector api.EventsSelector, err *meta.ErrBadRequest) {
				require.Nil(t, err)
				require.Equal(
					t,
					api.EventsSelector{
						ProjectID: "blue-book",
						Source:    "brigade.sh/cli",
						Qualifiers: api.Qualifiers{
							"foo": "bar",
							"bat": "baz",
						},
						Labels: map[string]string{
							"abc": "easy-as",
							"123": "do-rei-mei",
						},
						SourceState: map[string]string{
							"baby": "you",
							"and":  "me-girl",
						},
						Type: "exec",
						WorkerPhases: []api.WorkerPhase{
							api.WorkerPhasePending,
							api.WorkerPhaseStarting,
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
