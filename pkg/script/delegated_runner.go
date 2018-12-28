package script

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/Masterminds/kitt/progress"

	"github.com/Azure/brigade/pkg/brigade"
	"github.com/Azure/brigade/pkg/storage"
	"github.com/Azure/brigade/pkg/storage/kube"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var logPattern = regexp.MustCompile(`\[brigade:k8s\]\s[a-zA-Z0-9-]+/[a-zA-Z0-9-]+ phase \w+`)

const (
	waitTimeout = 5 * time.Minute
)

// NewDelegatedRunner returns a new Runner object with the provided Kubernetes Clientset and namespace
func NewDelegatedRunner(c *kubernetes.Clientset, namespace string) (*Runner, error) {
	app := &Runner{
		store:                kube.New(c, namespace),
		kc:                   c,
		namespace:            namespace,
		ScriptLogDestination: os.Stdout,
		RunnerLogDestination: os.Stdout,
	}
	return app, nil
}

// Runner represents an object that delegates running a script on Kubernetes
type Runner struct {
	store     storage.Store
	kc        kubernetes.Interface
	namespace string

	ScriptLogDestination io.Writer
	RunnerLogDestination io.Writer

	NoProgress bool
	Background bool
	Verbose    bool
}

// SendBuild creates and runs a given Brigade build
func (a *Runner) SendBuild(b *brigade.Build) error {
	if err := a.store.CreateBuild(b); err != nil {
		return err
	}

	podName := fmt.Sprintf("brigade-worker-%s", b.ID)

	if a.Background {
		fmt.Fprintf(a.RunnerLogDestination, "Build: %s, Worker: %s\n", b.ID, podName)
		return nil
	}
	fmt.Fprintf(a.RunnerLogDestination, "Event created. Waiting for worker pod named %q.\n", podName)

	if err := a.waitForWorker(b.ID); err != nil {
		return err
	}

	fmt.Fprintf(a.RunnerLogDestination, "Build: %s, Worker: %s\n", b.ID, podName)
	if err := a.podLog(podName, a.ScriptLogDestination); err != nil {
		return err
	}

	// Now that everything is complete, get the pod status. If the pod failed, exit 1.
	// Fortunately, the worker pod is marked "failed" if one of the jobs
	// in the build fails and the error isn't handled with a .catch().
	pod, err := a.kc.CoreV1().Pods(a.namespace).Get(podName, metav1.GetOptions{IncludeUninitialized: true})
	if err != nil {
		return err
	}

	podPhase := pod.Status.Phase
	if podPhase == v1.PodFailed || podPhase == v1.PodUnknown {
		return NewBuildFailure("build failed. (Build ID: %s)", b.ID)
	}

	return nil
}

// SendScript converts a script into a Brigade Build object and submits it
func (a *Runner) SendScript(projectName string, data []byte, event, commitish, ref string, payload []byte, logLevel string) error {

	projectID := brigade.ProjectID(projectName)
	if _, err := a.store.GetProject(projectID); err != nil {
		return fmt.Errorf("could not find the project %q: %s", projectName, err)
	}

	b := &brigade.Build{
		ProjectID: projectID,
		Type:      event,
		Provider:  "brigade-cli",
		Revision: &brigade.Revision{
			Commit: commitish,
			Ref:    ref,
		},
		Payload:  payload,
		Script:   data,
		LogLevel: logLevel,
	}
	return a.SendBuild(b)
}

// waitForWorker waits until the worker has started.
func (a *Runner) waitForWorker(buildID string) error {
	opts := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("heritage=brigade,component=build,build=%s", buildID),
	}
	req, err := a.kc.CoreV1().Pods(a.namespace).Watch(opts)
	if err != nil {
		return err
	}
	res := req.ResultChan()

	// Now we block until the Pod is ready
	timeout := time.After(2 * time.Minute)
	for {
		select {
		case e := <-res:
			if a.Verbose {
				d, _ := json.MarshalIndent(e.Object, "", "  ")
				fmt.Fprintf(a.RunnerLogDestination, "Event: %s\n %s\n", e.Type, d)
			}
			// If the pod is added or modified, check the phase and see if it is
			// running or complete.
			switch e.Type {
			case "DELETED":
				// This happens if a user directly kills the pod with kubectl.
				return fmt.Errorf("worker %s was just deleted unexpectedly", buildID)
			case "ADDED", "MODIFIED":
				pod := e.Object.(*v1.Pod)
				switch pod.Status.Phase {
				// Unhandled cases are Unknown and Pending, both of which should
				// cause the loop to spin.
				case "Running", "Succeeded":
					req.Stop()
					return nil
				case "Failed":
					req.Stop()
					return fmt.Errorf("pod failed to schedule: %s", pod.Status.Reason)
				}
			}
		case <-timeout:
			req.Stop()
			return fmt.Errorf("timeout waiting for build %s to start", buildID)
		}
	}
}

func (a *Runner) podLog(name string, w io.Writer) error {
	req := a.kc.CoreV1().Pods(a.namespace).GetLogs(name, &v1.PodLogOptions{Follow: true})

	readCloser, err := req.Timeout(waitTimeout).Stream()
	if err != nil {
		return err
	}
	defer readCloser.Close()

	if !a.NoProgress {
		progressLogs(w, readCloser)
	}

	_, err = io.Copy(w, readCloser)
	return err
}

// GetBuild retrieves a given build by id from storage
func (a *Runner) GetBuild(bid string) (*brigade.Build, error) {
	return a.store.GetBuild(bid)
}

func progressLogs(w io.Writer, r io.Reader) {
	scanner := bufio.NewScanner(r)
	last := []byte{}
	p := &progress.Indicator{
		Interval: 200 * time.Millisecond,
		Writer:   w,
		Frames: []string{
			"....",
			"=...",
			".=..",
			"..=.",
			"...=",
			"....",
			"...=",
			"..=.",
			".=..",
			"=...",
		},
	}
	started := false
	for scanner.Scan() {
		raw := scanner.Bytes()
		if string(raw) == string(last) && logPattern.Match(raw) {
			if started {
				continue
			}
			name := strings.Fields(string(raw))
			p.Start(name[len(name)-1])
			started = true
		} else {
			if started {
				p.Done("done")
				started = false
			}
			w.Write(raw)
			w.Write([]byte{'\n'})
		}
		last = raw
	}
}
