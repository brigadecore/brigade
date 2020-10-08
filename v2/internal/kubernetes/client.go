package kubernetes

import (
	"github.com/brigadecore/brigade/v2/internal/os"
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Client returns a new Kubernetes *kubernetes.Clientset.
func Client() (*kubernetes.Clientset, error) {
	masterURL := os.GetEnvVar("KUBE_MASTER", "")
	kubeConfigPath := os.GetEnvVar("KUBE_CONFIG", "")

	var cfg *rest.Config
	var err error
	if masterURL == "" && kubeConfigPath == "" {
		cfg, err = rest.InClusterConfig()
	} else {
		cfg, err = clientcmd.BuildConfigFromFlags(masterURL, kubeConfigPath)
	}
	if err != nil {
		return nil, errors.Wrap(
			err,
			"error getting kubernetes configuration",
		)
	}
	return kubernetes.NewForConfig(cfg)
}
