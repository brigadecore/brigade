package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/uswitch/brigade/pkg/brigade"
	"github.com/uswitch/brigade/pkg/storage"
	"github.com/uswitch/brigade/pkg/storage/kube"
)

var (
	runFile      string
	runEvent     string
	runPayload   string
	runCommitish string
)

const (
	defaultCommit = "master"
	kubeConfig    = "KUBECONFIG"
	waitTimeout   = 5 * time.Minute
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
	run.Flags().StringVarP(&runCommitish, "commit", "c", defaultCommit, "A VCS (git) commit version, tag, or branch")
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

		return a.send(proj, script)
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
}

func (a *scriptRunner) send(projectName string, data []byte) error {
	b := &brigade.Build{
		ProjectID: brigade.ProjectID(projectName),
		Type:      a.event,
		Provider:  "brigade-cli",
		Revision: &brigade.Revision{
			Ref: a.commit,
		},
		Payload: a.payload,
		Script:  data,
	}

	if err := a.store.CreateBuild(b); err != nil {
		return err
	}

	podName := fmt.Sprintf("brigade-worker-%s", b.ID)

	if err := a.waitForWorker(b.ID); err != nil {
		return err
	}

	fmt.Printf("Started build %s as %q\n", b.ID, podName)
	return a.podLog(podName, os.Stdout)
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

	_, err = io.Copy(w, readCloser)
	return err
}
