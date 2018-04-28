package apicache

import (
	"k8s.io/api/core/v1"
	"github.com/davecgh/go-spew/spew"
)

func (a *apiCache) GetPodsFilteredBy(labelSelectors map[string]string) []v1.Pod {

	var filteredSecrets []v1.Pod

	spew.Dump(a.podStore.List())

OuterLoop:
	for _, raw := range a.podStore.List() {

		secret, ok := raw.(*v1.Pod)
		if !ok {
			continue
		}

		// if the key doesn't exist on the secret or the expected value differs from the actual value, skip it
		for key, expected := range labelSelectors {
			actual, exists := secret.Labels[key]
			if !exists || actual != expected {
				continue OuterLoop
			}
		}

		filteredSecrets = append(filteredSecrets, *secret)
	}

	return filteredSecrets
}