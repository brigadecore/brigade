package kube

import (
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetStorageClassNames returns the names of the StorageClass instances in the cluster
func (s *store) GetStorageClassNames() ([]string, error) {
	scl, err := s.client.StorageV1().StorageClasses().List(meta.ListOptions{})
	if err != nil {
		return nil, err
	}

	var scss []string
	for _, sc := range scl.Items {
		scss = append(scss, sc.Name)
	}

	return scss, nil
}
