package js

import (
	"testing"
)

func TestSandbox_ExecAll(t *testing.T) {
	globals := map[string]interface{}{
		"github": map[string]string{
			"org":     "deis",
			"project": "acid",
		},
	}
	script1 := []byte(`console.log(github.org + "/" + github.project)`)
	script2 := []byte(`console.log(github.project+ "/" + github.org)`)

	s, err := New()
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range globals {
		s.Variable(k, v)
	}

	if err := s.ExecAll(script1, script2); err != nil {
		t.Fatalf("JavaScript failed to execute with error %s", err)
	}
}

func TestSandbox_ExecString(t *testing.T) {
	s, err := New()
	if err != nil {
		t.Fatal(err)
	}

	if err := s.ExecString(`console.log("hello there")`); err != nil {
		t.Fatal(err)
	}

	if err := s.ExecString(`invalid.f(`); err == nil {
		t.Fatal("exepected invalid JS to produce an error")
	}
}
