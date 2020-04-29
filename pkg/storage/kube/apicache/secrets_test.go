package apicache

import (
	"context"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/brigadecore/brigade/pkg/merge"
)

func TestSecretStore(t *testing.T) {

	client := fake.NewSimpleClientset()

	secretsSynced := make(chan struct{})
	podsSynced := make(chan struct{})
	merged := merge.Channels(secretsSynced, podsSynced)

	store := newSecretStore(client, "default", 1, secretsSynced)

	validLabels := map[string]string{
		"foo": "bar",
	}

	invalidLabels := map[string]string{
		"bar": "baz",
	}

	pod1 := v1.Pod{
		ObjectMeta: metaV1.ObjectMeta{
			Name: "pod1",
		},
	}

	secret1 := v1.Secret{
		ObjectMeta: metaV1.ObjectMeta{
			Labels: validLabels,
			Name:   "secret1",
		},
	}

	secret2 := v1.Secret{
		ObjectMeta: metaV1.ObjectMeta{
			Labels: invalidLabels,
			Name:   "secret2",
		},
	}

	if _, err := client.CoreV1().Secrets("default").Create(context.TODO(), &secret1, metaV1.CreateOptions{}); err != nil {
		t.Fatal(err)
	}

	if _, err := client.CoreV1().Secrets("default").Create(context.TODO(), &secret2, metaV1.CreateOptions{}); err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Millisecond)

	// inject a non pod object into the cache (which should get discarded)
	store.Add(&pod1)

	cache := apiCache{
		hasSyncedInitially: merged,
		client:             client,
		secretStore:        store,
		podStore:           newPodStore(client, "default", 1, podsSynced),
	}

	filteredPods, err := cache.GetSecretsFilteredBy(validLabels)
	if err != nil {
		t.Fatal(err)
	}
	if len(filteredPods) != 1 {
		t.Fatal("expected len(filtered pods) to be 1")
	}
}
