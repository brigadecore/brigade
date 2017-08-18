package kube

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// GetClient creates a config from the given master and kubeconfig
// location on disk, then creates a new kubernetes Clientset from that config
func GetClient(master, kubeConfigLocation string) (*kubernetes.Clientset, error) {
	// build the config from the master and kubeconfig location
	config, err := clientcmd.BuildConfigFromFlags(master, kubeConfigLocation)
	if err != nil {
		return nil, err
	}

	// creates the clientset
	return kubernetes.NewForConfig(config)
}
