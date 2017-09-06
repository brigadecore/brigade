package kube

import (
	"k8s.io/client-go/kubernetes"
)

// store represents a storage engine for a acid.Project.
type store struct {
	client    kubernetes.Interface
	namespace string
}

// New initializes a new storage backend.
func New(c kubernetes.Interface, namespace string) *store {
	return &store{c, namespace}
}
