package kubernetes

import (
	"testing"

	"github.com/brigadecore/brigade/v2/apiserver/internal/api"
	myk8s "github.com/brigadecore/brigade/v2/internal/kubernetes"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/kubernetes/fake"
)

func TestNewLogsStore(t *testing.T) {
	kubeClient := fake.NewSimpleClientset()
	store, ok := NewLogsStore(kubeClient).(*logsStore)
	require.True(t, ok)
	require.Same(t, kubeClient, store.kubeClient)
}

// TODO: This is very difficult, if not impossible, to test in isolation because
// the fake Kubernetes client (or a real one, for that matter) does not expose
// any way to store logs that can later be retrieved by the test cases. We'll
// have to make sure this behavior is well-covered by integration or e2e tests
// in the future.
func TestLogStoreStreamLogs(t *testing.T) {
	// require.Fail(t, "test me")
}

func TestPodNameFromSelector(t *testing.T) {
	const testEventID = "123456789"
	const testJobName = "italian"
	testCases := []struct {
		name            string
		selector        api.LogsSelector
		expectedPodName string
	}{
		{
			name:            "job not specified",
			selector:        api.LogsSelector{},
			expectedPodName: myk8s.WorkerPodName(testEventID),
		},
		{
			name: "job specified",
			selector: api.LogsSelector{
				Job: testJobName,
			},
			expectedPodName: myk8s.JobPodName(testEventID, testJobName),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			podName := podNameFromSelector(testEventID, testCase.selector)
			require.Equal(t, testCase.expectedPodName, podName)
		})
	}
}
