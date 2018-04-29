package kube

import (
	"k8s.io/client-go/kubernetes"
	"github.com/Azure/brigade/pkg/storage"
	"github.com/Azure/brigade/pkg/storage/kube/apicache"
	"time"
)

// store represents a storage engine for a brigade.Project.
type store struct {
	client    kubernetes.Interface
	namespace string
	apiCache  apicache.ApiCache
}

// Blocks until the ApiCache is populated, useful for testing
func (s *store)BlockUntilApiCacheSynced(waitUntil <- chan time.Time)bool{
	return s.apiCache.BlockUntilApiCacheSynced(waitUntil)
}

// New initializes a new storage backend.
func New(c kubernetes.Interface, namespace string) storage.Store {
	return &store{
		client:    c,
		namespace: namespace,
		apiCache: apicache.New(c,namespace,time.Duration(60) * time.Second),
	}
}
