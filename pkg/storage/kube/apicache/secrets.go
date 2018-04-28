package apicache

import (
	"k8s.io/api/core/v1"
)

// returns all secrets filtered by a label selector
// e.g. for 'heritage=brigade,component=build,project=%s'
// map[string]string{
//	"heritage":  "brigade",
//	"component": "build",
//	"project":   proj.ID,
// }
func (a *apiCache) GetSecretsFilteredBy(labelSelectors map[string]string) []v1.Secret {

	var filteredSecrets []v1.Secret

	OuterLoop:
	for _, raw := range a.secretStore.List() {

		secret, ok := raw.(*v1.Secret)
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
