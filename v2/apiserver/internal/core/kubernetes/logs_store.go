package kubernetes

import (
	"bufio"
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/brigadecore/brigade/v2/apiserver/internal/core"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	myk8s "github.com/brigadecore/brigade/v2/internal/kubernetes"
	"github.com/brigadecore/brigade/v2/internal/retries"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
)

// logsStore is a Kubernetes-based implementation of the core.LogsStore
// interface.
type logsStore struct {
	kubeClient kubernetes.Interface
}

// NewLogsStore returns a Kubernetes-based implementation of the core.LogsStore
// interface. It can stream logs directly from a Worker or Job's underlying pod.
// In practice, this is useful for very near-term retrieval of Worker and Job
// logs without incurring the latency inherent in other implementations that
// rely on a log aggregator having forwarded and stored log entries. This
// implementation will error, however, once the relevant pod has been deleted.
// Callers should be prepared to fall back on another implementation of the
// core.LogsStore interface, with the assumption that by the time a Worker's or
// Job's pod has been deleted, all of its logs have been aggregated and stored.
func NewLogsStore(kubeClient kubernetes.Interface) core.LogsStore {
	return &logsStore{
		kubeClient: kubeClient,
	}
}

func (l *logsStore) StreamLogs(
	ctx context.Context,
	project core.Project,
	event core.Event,
	selector core.LogsSelector,
	opts core.LogStreamOptions,
) (<-chan core.LogEntry, error) {
	podName := podNameFromSelector(event.ID, selector)

	req := l.kubeClient.CoreV1().Pods(project.Kubernetes.Namespace).GetLogs(
		podName,
		&v1.PodLogOptions{
			Container:  selector.Container,
			Timestamps: true,
			Follow:     opts.Follow,
		},
	)

	// The LogsService only would have called us for a Worker or Job that has
	// already moved past the PENDING and STARTING phases. So at this point, the
	// only two possibilities are that the Pod exists OR that it DID exist and has
	// already blinked out of existence after completion. If it's gone, we just
	// return a *meta.ErrNotFound and the LogsService will fall back to the cool
	// logs. If it exists, but the target container is still initializing, we
	// retry.
	var podLogs io.ReadCloser
	var err error
	if err = retries.ManageRetries(
		ctx,
		"waiting for container to be initialized",
		50, // A generous number of retries. Let the client hang up if they want.
		10*time.Second,
		func() (bool, error) {
			podLogs, err = req.Stream(ctx)
			if err != nil {
				if statusErr, ok := err.(*k8sErrors.StatusError); ok {
					if statusErr.Status().Code == http.StatusNotFound {
						return false, &meta.ErrNotFound{ // Don't retry
							Type: "Pod",
							ID:   podName,
						}
					}
					if strings.Contains(
						statusErr.Error(),
						"is waiting to start: PodInitializing",
					) || strings.Contains(
						statusErr.Error(),
						"is waiting to start: ContainerCreating",
					) {
						return true, nil // Retry
					}
				}
				// Something else is wrong
				return false, errors.Wrapf( // Don't retry
					err,
					"error opening log stream for pod %q in namespace %q",
					podName,
					project.Kubernetes.Namespace,
				)
			}
			return false, nil // We got what we wanted
		},
	); err != nil {
		return nil, err
	}

	logEntryCh := make(chan core.LogEntry)

	go func() {
		defer podLogs.Close()
		defer close(logEntryCh)
		buffer := bufio.NewReader(podLogs)
		for {
			logEntry := core.LogEntry{}
			logLine, err := buffer.ReadString('\n')
			if err == io.EOF {
				break
			}
			if len(logLine) == 0 {
				continue
			}
			// The last character should be a newline that we don't want, so let's
			// remove that
			logLine = logLine[:len(logLine)-1]
			logLineParts := strings.SplitN(logLine, " ", 2)
			if len(logLineParts) == 2 {
				timeStr := logLineParts[0]
				t, err := time.Parse(time.RFC3339, timeStr)
				if err == nil {
					logEntry.Time = &t
				}
				logEntry.Message = logLineParts[1]
			} else {
				logEntry.Message = logLine
			}
			select {
			case logEntryCh <- logEntry:
			case <-ctx.Done():
				return
			}
		}
		podLogs.Close()
	}()

	return logEntryCh, nil
}

func podNameFromSelector(eventID string, selector core.LogsSelector) string {
	if selector.Job == "" { // We want worker logs
		return myk8s.WorkerPodName(eventID)
	}
	return myk8s.JobPodName(eventID, selector.Job) // We want job logs
}
