package k8s

import (
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func Client() (kubernetes.Interface, error) {
	k8scfg := os.Getenv("KUBECONFIG")
	if k8scfg == "" {
		k8scfg = os.Getenv("HOME") + "/.kube/config"
	}

	var cfg *rest.Config
	var err error
	if _, err := os.Stat(k8scfg); err != nil {
		cfg, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	} else {
		cfg, err = clientcmd.BuildConfigFromFlags("", k8scfg)
		if err != nil {
			return nil, err
		}
	}
	k8s, err := kubernetes.NewForConfig(cfg)
	return k8s, err
}
