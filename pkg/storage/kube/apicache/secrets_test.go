package apicache

import (
	"testing"
	"time"

	"k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestSecretStore(t *testing.T) {

	client := fake.NewSimpleClientset()

	factory := secretStoreFactory{}
	store := factory.new(client, "default", 1, nil)

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

	_, err := client.CoreV1().Secrets("default").Create(&secret1)
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.CoreV1().Secrets("default").Create(&secret2)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Millisecond)

	// inject a non pod object into the cache (which should get discarded)
	store.Add(&pod1)

	cache := apiCache{
		client:      client,
		secretStore: store,
	}

	filteredPods := cache.GetSecretsFilteredBy(validLabels)
	if len(filteredPods) != 1 {
		t.Fatal("expected len(filtered pods) to be 1")
	}
}
