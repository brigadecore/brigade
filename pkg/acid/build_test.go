package acid

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/api/v1"
)

func TestNewBuildFromSecret(t *testing.T) {
	secret := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"build":   "#1",
				"project": "myproject",
				"commit":  "abc123",
			},
		},
		StringData: map[string]string{
			"event_type":     "foo",
			"event_provider": "bar",
		},
		Data: map[string][]byte{
			"payload": []byte("this is a payload"),
			"script":  []byte("ohai"),
		},
	}
	expectedBuild := &Build{
		ID:        "#1",
		ProjectID: "myproject",
		Commit:    "abc123",
		Type:      "foo",
		Provider:  "bar",
		Payload:   []byte("this is a payload"),
		Script:    []byte("ohai"),
	}
	build := NewBuildFromSecret(secret)
	if !reflect.DeepEqual(build, expectedBuild) {
		t.Errorf("build differs from expected build, got '%v', expected '%v'", build, expectedBuild)
	}
}
