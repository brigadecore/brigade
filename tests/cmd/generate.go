package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/google/go-github/github"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/deis/acid/pkg/storage/kube"
	"github.com/deis/acid/pkg/webhook"
)

var (
	kubeconfig string
	master     string
	namespace  string
)

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "absolute path to the kubeconfig file")
	flag.StringVar(&master, "master", "", "master url")
	flag.StringVar(&namespace, "namespace", os.Getenv("ACID_NAMESPACE"), "kubernetes namespace")
}

func getKubeClient() (*kubernetes.Clientset, error) {
	// creates the connection
	config, err := clientcmd.BuildConfigFromFlags(master, kubeconfig)
	if err != nil {
		return nil, err
	}

	// creates the clientset
	return kubernetes.NewForConfig(config)
}

func main() {
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "required arg: Git SHA")
		os.Exit(1)
	}
	commit := flag.Arg(0)

	eventType := "push"

	data, err := ioutil.ReadFile("./tests/testdata/test-repo-push.json")
	if err != nil {
		panic(err)
	}

	event, err := github.ParseWebHook(eventType, data)
	if err != nil {
		panic(err)
	}

	var repo string

	switch event := event.(type) {
	case *github.PushEvent:
		event.HeadCommit.ID = github.String(commit)
		repo = event.Repo.GetFullName()
	case *github.PullRequestEvent:
		event.PullRequest.Head.SHA = github.String(commit)
		repo = event.Repo.GetFullName()
	}

	out, err := json.MarshalIndent(event, "", "  ")
	if err != nil {
		panic(err)
	}

	clientset, err := getKubeClient()
	if err != nil {
		panic(err)
	}

	if namespace == "" {
		namespace = v1.NamespaceDefault
	}
	proj, err := kube.New(clientset, namespace).GetProject(repo)
	if err != nil {
		panic(err)
	}

	hmac := webhook.SHA1HMAC([]byte(proj.SharedSecret), out)

	ioutil.WriteFile("./tests/testdata/test-repo-generated.json", out, 0755)
	ioutil.WriteFile("./tests/testdata/test-repo-generated.hash", []byte(hmac), 0755)

	fmt.Fprintln(os.Stdout, string(out))
	fmt.Fprintln(os.Stdout, hmac)
}
