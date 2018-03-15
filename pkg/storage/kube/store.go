package kube

import (
	"k8s.io/client-go/kubernetes"

	"github.com/Azure/brigade/pkg/storage"
)

// store represents a storage engine for a brigade.Project.
type store struct {
	client    kubernetes.Interface
	namespace string
}

// New initializes a new storage backend.
func New(c kubernetes.Interface, namespace string) storage.Store {
	return &store{c, namespace}
}
