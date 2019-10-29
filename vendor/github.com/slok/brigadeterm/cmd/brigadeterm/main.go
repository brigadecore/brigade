package main

import (
	"fmt"
	"os"

	"github.com/brigadecore/brigade/pkg/storage/kube"
	"github.com/rivo/tview"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/slok/brigadeterm/pkg/controller"
	"github.com/slok/brigadeterm/pkg/service/brigade"
	"github.com/slok/brigadeterm/pkg/ui"
)

const (
	brigadeVersionFMT = "brigadeterm %s"
	kubeCliQPS        = 50
	kubeCliBurst      = 50
)

var (
	// Version is the app version.
	Version = "dev"
)

// Main is the main package.
type Main struct {
	flags *cmdFlags
}

// NewMain returns a new main application.
func NewMain() (*Main, error) {
	flags, err := newCmdFlags()
	if err != nil {
		return nil, err
	}

	return &Main{
		flags: flags,
	}, nil
}

// Run will run the main application.
func (m *Main) Run() error {
	if m.flags.showVersion {
		m.printVersion()
		return nil
	}
	// Create external dependencies.
	k8scli, err := m.createKubernetesClients()
	if err != nil {
		return err
	}
	brigadek8s := kube.New(k8scli, m.flags.brigadeNamespace)

	// Create services
	brigadeService := brigade.NewService(brigadek8s)

	// Create controller.
	var uictrl controller.Controller
	if m.flags.fake {
		uictrl = controller.NewFakeController()
	} else {
		uictrl = controller.NewController(brigadeService)
	}

	// Create the terminal app.
	app := tview.NewApplication()

	index := ui.NewIndex(m.flags.reloadInterval, uictrl, app)

	return index.Render()
}

func (m *Main) createKubernetesClients() (kubernetes.Interface, error) {
	config, err := m.loadKubernetesConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

// loadKubernetesConfig loads kubernetes configuration based on flags.
func (m *Main) loadKubernetesConfig() (*rest.Config, error) {
	loader := &clientcmd.ClientConfigLoadingRules{
		ExplicitPath: m.flags.kubeConfig,
	}
	overrides := &clientcmd.ConfigOverrides{
		CurrentContext: m.flags.kubeContext,
	}
	cfg, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loader, overrides).ClientConfig()
	if err != nil {
		return nil, err
	}

	// Set better client rate limits.
	cfg.QPS = kubeCliQPS
	cfg.Burst = kubeCliBurst

	return cfg, nil
}

// printVersion prints the version of the app.
func (m *Main) printVersion() {
	fmt.Fprintf(os.Stdout, brigadeVersionFMT, Version)
}

func main() {
	m, err := NewMain()
	if err != nil {
		exitWithError(err)
	}

	if err := m.Run(); err != nil {
		exitWithError(err)
	}
	os.Exit(0)
}

func exitWithError(err error) {
	fmt.Fprintf(os.Stderr, err.Error()+"\n")
	os.Exit(1)
}
