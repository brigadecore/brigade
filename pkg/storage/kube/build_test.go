package kube

import (
	"reflect"
	"testing"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Azure/brigade/pkg/brigade"
)

func TestNewBuildFromSecret(t *testing.T) {
	secret := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"build":   "#1",
				"project": "myproject",
			},
		},
		Data: map[string][]byte{
			"event_type":     []byte("foo"),
			"event_provider": []byte("bar"),
			"payload":        []byte("this is a payload"),
			"script":         []byte("ohai"),
			"commit_id":      []byte("abc123"),
			"commit_ref":     []byte("refs/heads/master"),
		},
	}
	expectedBuild := &brigade.Build{
		ID:        "#1",
		ProjectID: "myproject",
		Revision: &brigade.Revision{
			Commit: "abc123",
			Ref:    "refs/heads/master",
		},
		Type:     "foo",
		Provider: "bar",
		Payload:  []byte("this is a payload"),
		Script:   []byte("ohai"),
	}
	build := NewBuildFromSecret(secret)
	if !reflect.DeepEqual(build, expectedBuild) {
		t.Errorf("build differs from expected build, got '%v', expected '%v'", build, expectedBuild)
	}
}
