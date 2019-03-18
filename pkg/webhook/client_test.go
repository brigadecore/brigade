package webhook

import (
	"testing"

	"github.com/brigadecore/brigade/pkg/brigade"
)

func TestGHClient(t *testing.T) {
	gh := brigade.Github{
		Token:     "totallyFake",
		BaseURL:   "http://example.com/base/",
		UploadURL: "http://example.com/upload/",
	}

	c, err := ghClient(gh)
	if err != nil {
		t.Fatal(err)
	}

	if c.BaseURL.String() != gh.BaseURL {
		t.Errorf("Expected %q, got %q", c.BaseURL.String(), gh.BaseURL)
	}
	if c.UploadURL.String() != gh.UploadURL {
		t.Errorf("Expected %q, got %q", c.UploadURL.String(), gh.UploadURL)
	}
}
