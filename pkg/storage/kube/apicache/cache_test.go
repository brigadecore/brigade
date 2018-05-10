package apicache

import (
	"testing"
	"time"

	"k8s.io/client-go/kubernetes/fake"
)

func TestApiCache(t *testing.T) {

	client := fake.NewSimpleClientset()

	cache := New(client, "default", time.Millisecond*time.Duration(500))
	if cache == nil {
		t.Fatal("expected cache not to ne nil")
	}

	syncedWithinOneSecond := cache.BlockUntilAPICacheSynced(time.After(time.Second))
	if !syncedWithinOneSecond {
		t.Fatal("expected to sync within one second")
	}

	syncedWithoutTimeout := cache.BlockUntilAPICacheSynced(nil)
	if !syncedWithoutTimeout {
		t.Fatal("expected to sync without timeout")
	}
}

func TestApiCacheBlockUntilApiCacheSynced(t *testing.T) {

	client := fake.NewSimpleClientset()

	cache := New(client, "default", time.Millisecond*time.Duration(500))
	if cache == nil {
		t.Fatal("expected cache not to ne nil")
	}

	syncedAfterZeroTime := cache.BlockUntilAPICacheSynced(time.After(0))
	if syncedAfterZeroTime {
		t.Fatal("expected to sync within one second")
	}
}
