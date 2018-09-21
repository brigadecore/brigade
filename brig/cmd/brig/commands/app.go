package commands

import (
	"github.com/spf13/cobra"

	// Kube client doesn't support all auth providers by default.
	// this ensures we include all backends supported by the client.
	"k8s.io/client-go/kubernetes"
	// auth is a side-effect import for Client-Go
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const mainUsage = `Interact with the Brigade cluster service.

Brigade is a tool for scripting cluster workflows, and 'brig' is the command
line client for interacting with Brigade.

The most common use for this tool is to send a Brigade JavaScript file to the
cluster for execution. This is done with the 'brigade run' command.

	$ brig run -f my.js my-project

But the 'brig' command can also be used for learning about projects and
builds as well.

By default, Brigade learns about your Kubernetes cluster by inspect the $KUBECONFIG
environment variable.
`

var (
	globalNamespace   string
	globalKubeConfig  string
	globalKubeContext string
	globalVerbose     bool
)

func init() {
	f := Root.PersistentFlags()
	f.StringVarP(&globalNamespace, "namespace", "n", "default", "The Kubernetes namespace for Brigade")
	f.StringVar(&globalKubeConfig, "kubeconfig", "", "The path to a KUBECONFIG file, overrides $KUBECONFIG.")
	f.StringVar(&globalKubeContext, "kube-context", "", "The name of the kubeconfig context to use.")
	f.BoolVarP(&globalVerbose, "verbose", "v", false, "Turn on verbose output")
}

// Root is the root command
var Root = &cobra.Command{
	Use:   "brig",
	Short: "The Brigade client",
	Long:  mainUsage,
}

// kubeClient returns a Kubernetes clientset.
func kubeClient() (*kubernetes.Clientset, error) {
	cfg, err := getKubeConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(cfg)
}

// getKubeConfig returns a Kubernetes client config.
func getKubeConfig() (*rest.Config, error) {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	rules.DefaultClientConfig = &clientcmd.DefaultClientConfig
	rules.ExplicitPath = globalKubeConfig

	overrides := &clientcmd.ConfigOverrides{
		ClusterDefaults: clientcmd.ClusterDefaults,
		CurrentContext:  globalKubeContext,
	}

	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides).ClientConfig()
}
