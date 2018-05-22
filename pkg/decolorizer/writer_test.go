package decolorizer

import (
	"bytes"
	"testing"
)

func TestDecolorizeWriter(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{
			input:  "\033",
			expect: "\x1b",
		},
		{
			input:  "\033[31mHello\033[0m World",
			expect: "Hello World",
		},
		{
			input:  "\033[31;1mHello\033[0m World",
			expect: "Hello World",
		},
		{
			input:  "\033[38;2;255;82;197;48;2;155;106;0mHello\033[0m World",
			expect: "Hello World",
		},
	}

	for i, tt := range tests {
		out := bytes.NewBuffer(nil)
		decol := New(out)
		decol.Write([]byte(tt.input))
		if got := out.String(); got != tt.expect {
			t.Errorf("For test %d, expected %q but got %q", i, tt.expect, got)
		}
	}
}
