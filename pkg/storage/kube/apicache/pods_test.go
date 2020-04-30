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

func TestPodStore(t *testing.T) {

	client := fake.NewSimpleClientset()

	secretsSynced := make(chan struct{})
	podsSynced := make(chan struct{})
	merged := merge.Channels(secretsSynced, podsSynced)

	store := newPodStore(client, "default", 1, podsSynced)

	validLabels := map[string]string{
		"foo": "bar",
	}

	invalidLabels := map[string]string{
		"bar": "baz",
	}

	pod1 := v1.Pod{
		ObjectMeta: metaV1.ObjectMeta{
			Labels: validLabels,
			Name:   "pod1",
		},
	}

	pod2 := v1.Pod{
		ObjectMeta: metaV1.ObjectMeta{
			Labels: invalidLabels,
			Name:   "pod2",
		},
	}

	secret1 := v1.Secret{
		ObjectMeta: metaV1.ObjectMeta{
			Name: "secret1",
		},
	}

	if _, err := client.CoreV1().Pods("default").Create(context.TODO(), &pod1, metaV1.CreateOptions{}); err != nil {
		t.Fatal(err)
	}

	if _, err := client.CoreV1().Pods("default").Create(context.TODO(), &pod2, metaV1.CreateOptions{}); err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Millisecond)

	// inject a non pod object into the cache (which should get discarded)
	store.Add(&secret1)

	cache := apiCache{
		hasSyncedInitially: merged,
		client:             client,
		podStore:           store,
		secretStore:        newSecretStore(client, "default", 1, secretsSynced),
	}

	filteredPods, err := cache.GetPodsFilteredBy(validLabels)
	if err != nil {
		t.Fatal(err)
	}
	if len(filteredPods) != 1 {
		t.Fatal("expected len(filtered pods) to be 1")
	}
}
