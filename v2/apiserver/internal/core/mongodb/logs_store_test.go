package mongodb

import (
	"testing"

	"github.com/brigadecore/brigade/v2/apiserver/internal/core"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

// TODO: This is very difficult to test in isolation. The implementation of the
// mongo-driver library doesn't lend itself to being easily mocked out. It has
// been mocked out, with great difficulty, elsewhere in our unit test suite, but
// streaming logs from the database involves the use of a "live," tailable
// cursor-- the nature of which is that when its current batch of data is
// exhausted, a subsequent batch is retrieved from the database. This is so
// enormously complex to mock out that I (krancour) am choosing to defer this,
// possibly indefinitely. We'll have to make sure this behavior is well-covered
// by integration or e2e tests in the future.
func TestLogStoreStreamLogs(t *testing.T) {
	// require.Fail(t, "test me")
}

func TestCriteriaFromSelector(t *testing.T) {
	const testEventID = "123456789"
	const testJobName = "italian"
	const testContainerName = "foo"
	testCases := []struct {
		name             string
		selector         core.LogsSelector
		expectedCriteria bson.M
	}{
		{
			name: "job not specified",
			selector: core.LogsSelector{
				// The service layer will ALWAYS have set this field if it wasn't set
				// already.
				Container: testContainerName,
			},
			expectedCriteria: bson.M{
				"event":     testEventID,
				"component": "worker",
				"container": testContainerName,
			},
		},
		{
			name: "job specified",
			selector: core.LogsSelector{
				Job: testJobName,
				// The service layer will ALWAYS have set this field if it wasn't set
				// already.
				Container: testContainerName,
			},
			expectedCriteria: bson.M{
				"event":     testEventID,
				"component": "job",
				"job":       testJobName,
				"container": testContainerName,
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			criteria := criteriaFromSelector(testEventID, testCase.selector)
			require.Equal(t, testCase.expectedCriteria, criteria)
		})
	}
}
