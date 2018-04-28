package apicache

import (
	"testing"
	"k8s.io/client-go/kubernetes/fake"
	"time"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestApiCache(t *testing.T){

	labels := map[string]string{
		"foo": "bar",
	}

	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Labels: labels,
		},
		Spec: v1.PodSpec{
			Hostname: "foo",
		},
		Status: v1.PodStatus{

		},
	}

	client := fake.NewSimpleClientset(&pod)
	// unexpectedly the api cache stays empty
	apiCache := New(client,"default",time.Second * time.Duration(30))

	pods := apiCache.GetPodsFilteredBy(labels)
	if len(pods) != 1 {
		t.Fatalf("expected 1 but found %d",len(pods))
	}
}