package kube

import (
	"time"

	"k8s.io/client-go/kubernetes"

	"github.com/Azure/brigade/pkg/storage"
	"github.com/Azure/brigade/pkg/storage/kube/apicache"
)

// store represents a storage engine for a brigade.Project.
type store struct {
	client    kubernetes.Interface
	namespace string
	apiCache  apicache.APICache
}

// New initializes a new storage backend.
func New(c kubernetes.Interface, namespace string) storage.Store {
	return &store{
		client:    c,
		namespace: namespace,
		apiCache:  apicache.New(c, namespace, time.Duration(60)*time.Second),
	}
}
