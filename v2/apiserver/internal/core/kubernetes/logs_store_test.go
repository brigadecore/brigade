package kubernetes

import (
	"testing"

	"github.com/brigadecore/brigade/v2/apiserver/internal/core"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/kubernetes/fake"
)

func TestNewLogsStore(t *testing.T) {
	kubeClient := fake.NewSimpleClientset()
	store := NewLogsStore(kubeClient)
	require.Same(t, kubeClient, store.(*logsStore).kubeClient)
}

// TODO: This is very difficult, if not impossible, to test in isolation because
// the fake Kubernetes client (or a real one, for that matter) does not expose
// any way to store logs that can later be retrieved by the test cases. We'll
// have to make sure this behavior is well-covered by integration or e2e tests
// in the future.
func TestLogStoreStreamLogs(t *testing.T) {
	// require.Fail(t, "test me")
}

func TestCriteriaFromSelector(t *testing.T) {
	const testEventID = "123456789"
	testCases := []struct {
		name                  string
		selector              core.LogsSelector
		expectedPodName       string
		expectedContainerName string
	}{
		{
			name:                  "job not specified, container not specified",
			selector:              core.LogsSelector{},
			expectedPodName:       "worker-123456789",
			expectedContainerName: "worker",
		},
		{
			name: "job not specified, container specified",
			selector: core.LogsSelector{
				Container: "helper",
			},
			expectedPodName:       "worker-123456789",
			expectedContainerName: "helper",
		},
		{
			name: "job specified, container not specified",
			selector: core.LogsSelector{
				Job: "italian",
			},
			expectedPodName:       "job-123456789-italian",
			expectedContainerName: "italian",
		},
		{
			name: "job specified, container specified",
			selector: core.LogsSelector{
				Job:       "italian",
				Container: "helper",
			},
			expectedPodName:       "job-123456789-italian",
			expectedContainerName: "helper",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			podName, containerName :=
				criteriaFromSelector(testEventID, testCase.selector)
			require.Equal(t, testCase.expectedPodName, podName)
			require.Equal(t, testCase.expectedContainerName, containerName)
		})
	}
}
