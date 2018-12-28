package apicache

import (
	"testing"
	"time"

	"k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

func TestNewListStoreWithoutHasSyncedChan(t *testing.T) {
	client := fake.NewSimpleClientset()

	store := newListStore(client, storeConfig{
		resource:     "secrets",
		namespace:    "default",
		resyncPeriod: time.Millisecond * time.Duration(100),
		expectedType: &v1.Secret{},
		listFunc: func(client kubernetes.Interface, namespace string, options metaV1.ListOptions) (runtime.Object, error) {
			return client.CoreV1().Secrets(namespace).List(options)
		},
		watchFunc: func(client kubernetes.Interface, namespace string, options metaV1.ListOptions) (watch.Interface, error) {
			return client.CoreV1().Secrets(namespace).Watch(options)
		},
	}, nil)

	if store == nil {
		t.Fatal("expected store to not be nil when hasSynced is nil")
	}
}

func TestNewListStoreInvokeWatchFunctions(t *testing.T) {

	client := fake.NewSimpleClientset()

	hasSynced := make(chan struct{})
	store := newListStore(client, storeConfig{
		resource:     "secrets",
		namespace:    "default",
		resyncPeriod: 1,
		expectedType: &v1.Secret{},
		listFunc: func(client kubernetes.Interface, namespace string, options metaV1.ListOptions) (runtime.Object, error) {
			return client.CoreV1().Secrets(namespace).List(options)
		},
		watchFunc: func(client kubernetes.Interface, namespace string, options metaV1.ListOptions) (watch.Interface, error) {
			return client.CoreV1().Secrets(namespace).Watch(options)
		},
	}, hasSynced)

	if store == nil {
		t.Fatal("expected store to not be nil")
	}

	// cover the watch functions
	secret := v1.Secret{
		ObjectMeta: metaV1.ObjectMeta{
			Name: "fooBar",
		},
		StringData: map[string]string{
			"foo": "bar",
		},
	}

	created, err := client.CoreV1().Secrets("default").Create(&secret)
	if err != nil {
		t.Fatal(err)
	}

	<-hasSynced

	if len(store.List()) != 1 {
		t.Fatal("expected store to contain one object")
	}

	if _, err := client.CoreV1().Secrets("default").Update(created); err != nil {
		t.Fatal(err)
	}

	<-hasSynced

	if err := client.CoreV1().Secrets("default").Delete(created.Name, nil); err != nil {
		t.Fatal(err)
	}
}

func TestStringMapsMatch(t *testing.T) {

	expected := map[string]string{
		"foo": "bar",
		"bar": "baz",
	}

	keyMissing := map[string]string{
		"foo": "bar",
	}

	invalidValue := map[string]string{
		"foo": "bar",
		"bar": "bar",
	}

	if !stringMapsMatch(expected, expected) {
		t.Fatal("expected maps to match")
	}

	if stringMapsMatch(keyMissing, expected) {
		t.Fatal("expected not to match because key is missing")
	}

	if stringMapsMatch(invalidValue, expected) {
		t.Fatal("expected not to match because one value is invalid")
	}
}
