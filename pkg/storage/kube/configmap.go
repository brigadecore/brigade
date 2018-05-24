package kube

import (
	"errors"
	"log"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetScript retrieves brigade.js script from storage
func (s *store) GetScript(name string) (string, error) {
	if name == "" {
		return "", errors.New("Missing config map name")
	}
	configMap, err := s.client.CoreV1().ConfigMaps(s.namespace).Get(name, meta.GetOptions{})
	if err != nil {
		log.Printf("Error retrieving config map %s: %s", name, err)
		return "", err
	}
	script := configMap.Data["brigade.js"]
	return script, nil
}
