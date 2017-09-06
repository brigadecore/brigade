package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"

	"github.com/deis/acid/pkg/acid"
	"github.com/deis/acid/pkg/storage"
	"github.com/deis/acid/pkg/storage/kube"
)

const usage = "lsd [-f FILE] PROJECT_NAME"

var (
	file      string
	namespace string
)

const (
	kubeConfig = "KUBECONFIG"
)

func init() {
	flag.StringVar(&file, "file", "acid.js", "the script to run")
	flag.StringVar(&namespace, "ns", "default", "the Kubernetes namespace of Acid")
}

func main() {

	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		bail("Expected 'lsd [--file FILENAME] PROJECT_NAME'")
	}

	proj := args[0]

	script, err := ioutil.ReadFile(file)
	if err != nil {
		bail(err.Error())
	}

	a, err := NewApp()
	if err != nil {
		bail(err.Error())
	}

	if err := a.send(proj, script); err != nil {
		bail(err.Error())
	}
}

func bail(msg string) {
	fmt.Fprintln(os.Stderr, usage)
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}

func NewApp() (*App, error) {
	c, err := kube.GetClient("", os.Getenv(kubeConfig))
	if err != nil {
		return nil, err
	}

	app := &App{
		store: kube.New(c, namespace),
		kc:    c,
	}
	return app, nil
}

type App struct {
	store storage.Store
	kc    kubernetes.Interface
}

func (a *App) send(projectName string, data []byte) error {
	b := &acid.Build{
		ProjectID: acid.ProjectID(projectName),
		Type:      "exec",
		Provider:  "lsd",
		Commit:    "master",
		Payload:   []byte{},
		Script:    data,
	}

	if err := a.store.CreateBuild(b); err != nil {
		return err
	}

	podName := "acid-worker-" + b.ID + "-master"

	// This is a hack to give the scheduler time to create the resource.
	time.Sleep(3 * time.Second)

	fmt.Printf("Started %s\n", podName)
	if err := a.podLog(podName, os.Stdout); err != nil {
		return err
	}

	return nil
}

func (a *App) podLog(name string, w io.Writer) error {
	req := a.kc.CoreV1().Pods(namespace).GetLogs(name, &v1.PodLogOptions{Follow: true})

	readCloser, err := req.Stream()
	if err != nil {
		return err
	}
	defer readCloser.Close()

	_, err = io.Copy(w, readCloser)
	return err
}
