package commands

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/Masterminds/kitt/progress"
	"github.com/spf13/cobra"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/Azure/brigade/pkg/brigade"
	"github.com/Azure/brigade/pkg/storage"
	"github.com/Azure/brigade/pkg/storage/kube"
)

var (
	runFile       string
	runEvent      string
	runPayload    string
	runCommitish  string
	runRef        string
	runNoProgress bool
	runExitCode   bool
)

var logPattern = regexp.MustCompile("\\[brigade:k8s\\]\\s[a-zA-Z0-9-]+/[a-zA-Z0-9-]+ phase \\w+")

const (
	defaultRef  = "master"
	kubeConfig  = "KUBECONFIG"
	waitTimeout = 5 * time.Minute
)

const runUsage = `Send a Brigade JS file to the server.

This sends a file into the cluster and waits for it to complete. It accepts
a project name or project ID.

	$ brig run deis/empty-testbed

When no JS file is supplied, the project will be checked for a brigade.js file
in the associated repository.

To send a local JS file to the server, use the '-f' flag:

	$ brig run -f my.js deis/empty-testbed

While specifying an event is possible, use caution. Mny events expect a
particular payload.
`

func init() {
	run.Flags().StringVarP(&runFile, "file", "f", "", "The JavaScript file to execute")
	run.Flags().StringVarP(&runEvent, "event", "e", "exec", "The name of the event to fire")
	run.Flags().StringVarP(&runPayload, "payload", "p", "", "The path to a payload file")
	run.Flags().StringVarP(&runCommitish, "commit", "c", "", "A VCS (git) commit")
	run.Flags().StringVarP(&runRef, "ref", "r", defaultRef, "A VCS (git) version, tag, or branch")
	run.Flags().BoolVar(&runNoProgress, "no-progress", false, "Disable progress meter")
	run.Flags().BoolVar(&runExitCode, "exit-code", false, "Return a failing exit code if build fails")
	Root.AddCommand(run)
}

var run = &cobra.Command{
	Use:   "run PROJECT",
	Short: "Run a brigade.js file",
	Long:  runUsage,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("Project name required")
		}
		proj := args[0]

		var script []byte
		if len(runFile) > 0 {
			var err error
			if script, err = ioutil.ReadFile(runFile); err != nil {
				return err
			}
		}

		a, err := newScriptRunner()
		if err != nil {
			return err
		}

		err = a.send(proj, script)
		if err == nil {
			return nil
		}

		// If err is a BuildFailure, then we don't want Cobra to print the Usage
		// instructions on failure, since it's a pipeline issue and not a CLI issue.
		_, ok := err.(BuildFailure)
		if ok {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
		}

		return err
	},
}

func newScriptRunner() (*scriptRunner, error) {
	c, err := kube.GetClient("", kubeConfigPath())
	if err != nil {
		return nil, err
	}

	app := &scriptRunner{
		store:  kube.New(c, globalNamespace),
		kc:     c,
		event:  runEvent,
		commit: runCommitish,
		ref:    runRef,
	}
	if len(runPayload) > 0 {
		data, err := ioutil.ReadFile(runPayload)
		if err != nil {
			return nil, err
		}
		app.payload = data
	}
	return app, nil
}

type scriptRunner struct {
	store   storage.Store
	kc      kubernetes.Interface
	payload []byte
	event   string
	commit  string
	ref     string
}

func (a *scriptRunner) send(projectName string, data []byte) error {

	projectID := brigade.ProjectID(projectName)
	if _, err := a.store.GetProject(projectID); err != nil {
		return fmt.Errorf("could not find the project %q: %s", projectName, err)
	}

	b := &brigade.Build{
		ProjectID: projectID,
		Type:      a.event,
		Provider:  "brigade-cli",
		Revision: &brigade.Revision{
			Commit: a.commit,
			Ref:    a.ref,
		},
		Payload: a.payload,
		Script:  data,
	}

	if err := a.store.CreateBuild(b); err != nil {
		return err
	}

	podName := fmt.Sprintf("brigade-worker-%s", b.ID)

	fmt.Printf("Event created. Waiting for worker pod named %q.\n", podName)

	if err := a.waitForWorker(b.ID); err != nil {
		return err
	}

	fmt.Printf("Started build %s as %q\n", b.ID, podName)
	if err := a.podLog(podName, os.Stdout); err != nil {
		return err
	}

	// If the --exit-code flag is not set, at this point brig has completed without errors
	// and we can return. Note that if even if the build failed, brig will still exit 0.
	if !runExitCode {
		return nil
	}

	// Now that everything is complete, get the pod status. If the pod failed, exit 1.
	// Fortunately, the worker pod is marked "failed" if one of the jobs
	// in the build fails and the error isn't handled with a .catch().
	pod, err := a.kc.CoreV1().Pods(globalNamespace).Get(podName, metav1.GetOptions{IncludeUninitialized: true})
	if err != nil {
		return err
	}
	podPhase := pod.Status.Phase
	if podPhase == v1.PodFailed || podPhase == v1.PodUnknown {
		return NewBuildFailure("build failed. (Build ID: %s)", b.ID)
	}

	return nil
}

// waitForWorker waits until the worker has started.
func (a *scriptRunner) waitForWorker(buildID string) error {
	opts := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("heritage=brigade,component=build,build=%s", buildID),
	}
	req, err := a.kc.CoreV1().Pods(globalNamespace).Watch(opts)
	if err != nil {
		return err
	}
	res := req.ResultChan()

	// Now we block until the Pod is ready
	timeout := time.After(2 * time.Minute)
	for {
		select {
		case e := <-res:
			if globalVerbose {
				d, _ := json.MarshalIndent(e.Object, "", "  ")
				fmt.Printf("Event: %s\n %s\n", e.Type, d)
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

func (a *scriptRunner) podLog(name string, w io.Writer) error {
	req := a.kc.CoreV1().Pods(globalNamespace).GetLogs(name, &v1.PodLogOptions{Follow: true})

	readCloser, err := req.Timeout(waitTimeout).Stream()
	if err != nil {
		return err
	}
	defer readCloser.Close()

	if !runNoProgress {
		progressLogs(w, readCloser)
	}

	_, err = io.Copy(w, readCloser)
	return err
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
