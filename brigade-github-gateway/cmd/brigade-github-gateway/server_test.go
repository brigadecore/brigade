package main

import "testing"

func TestAuthors(t *testing.T) {
	expand := "a,b,c"

	a := authors{}
	a.Set(expand)
	if len(a) != 3 {
		t.Fatal("expected three authors")
	}
	for i, item := range []string{"A", "B", "C"} {
		if item != a[i] {
			t.Errorf("index %d: expected %s, got %s", i, item, a[i])
		}
	}

	collapse := authors{"TASHTEGO", "QUEEQUEG", "ISHMAEL"}
	expect := "TASHTEGO,QUEEQUEG,ISHMAEL"
	if got := collapse.String(); expect != got {
		t.Errorf("Expected %q, got %q", expect, got)
	}
}
