package commands

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/Azure/brigade/pkg/brigade"
	"github.com/Azure/brigade/pkg/storage"
	"github.com/Azure/brigade/pkg/storage/kube"
)

var (
	runFile    string
	runEvent   string
	runPayload string
)

const (
	kubeConfig   = "KUBECONFIG"
	waitTimeout  = 5 * time.Minute
	initialDelay = 5 * time.Second
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
	run.Flags().StringVarP(&runFile, "file", "f", "./brigade.js", "The JavaScript file to execute")
	run.Flags().StringVarP(&runEvent, "event", "e", "exec", "The name of the event to fire")
	run.Flags().StringVarP(&runPayload, "payload", "p", "", "The path to a payload file")
	// TODO: add support for specifying payload and event type
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
		store: kube.New(c, globalNamespace),
		kc:    c,
		event: runEvent,
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
}

func (a *scriptRunner) send(projectName string, data []byte) error {
	b := &brigade.Build{
		ProjectID: brigade.ProjectID(projectName),
		Type:      a.event,
		Provider:  "brigade-cli",
		Commit:    "master",
		Payload:   a.payload,
		Script:    data,
	}

	if err := a.store.CreateBuild(b); err != nil {
		return err
	}

	podName := "brigade-worker-" + b.ID + "-master"

	// This is a hack to give the scheduler time to create the resource.
	time.Sleep(initialDelay)

	fmt.Printf("Started %s\n", podName)
	if err := a.podLog(podName, os.Stdout); err != nil {
		return err
	}

	return nil
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
