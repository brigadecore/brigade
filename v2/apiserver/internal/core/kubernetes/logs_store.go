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
	podLogs, err := req.Stream(ctx)
	if err != nil {
		if statusErr, ok := err.(*k8sErrors.StatusError); ok {
			if statusErr.Status().Code == http.StatusNotFound {
				err = &meta.ErrNotFound{
					Type: "Pod",
					ID:   podName,
				}
			}
		}
		return nil, errors.Wrapf(
			err,
			"error opening log stream for pod %q in namespace %q",
			podName,
			project.Kubernetes.Namespace,
		)
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
