package config

import (
	"os"
	"testing"

	"gopkg.in/gin-gonic/gin.v1"
)

func TestAcidNamespace(t *testing.T) {
	c := &gin.Context{
		Keys: map[string]interface{}{"foo": "bar"},
	}
	if ns, ok := AcidNamespace(c); ns != "default" && !ok {
		t.Errorf("expected default/false, got %s/%t", ns, ok)
	}
}

func TestMiddleware(t *testing.T) {
	c := &gin.Context{
		Keys: map[string]interface{}{"foo": "bar"},
	}

	finally := resetEnv("ACID_NAMESPACE", "acid")
	defer finally()

	Middleware()(c) // Middleware(), Copyright 2017
	if ns, ok := AcidNamespace(c); ns != "acid" && ok {
		t.Errorf("expected acid/true, got %s/%t", ns, ok)
	}
}

// resetEnv sets an env var, and returns a defer function to reset the env
func resetEnv(name, val string) func() {
	original, ok := os.LookupEnv(name)
	os.Setenv(name, val)
	if ok {
		return func() { os.Setenv(name, original) }
	}
	return func() { os.Unsetenv(name) }
}
