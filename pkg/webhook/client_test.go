package webhook

import (
	"testing"

	"github.com/brigadecore/brigade/pkg/brigade"
)

func TestGHClient(t *testing.T) {
	gh := brigade.Github{
		Token:     "totallyFake",
		BaseURL:   "http://example.com/base/api/v3/",
		UploadURL: "http://example.com/upload/api/v3/",
	}

	c, err := ghClient(gh)
	if err != nil {
		t.Fatal(err)
	}

	if c.BaseURL.String() != gh.BaseURL {
		t.Errorf("Expected %q, got %q", gh.BaseURL, c.BaseURL.String())
	}
	if c.UploadURL.String() != gh.UploadURL {
		t.Errorf("Expected %q, got %q", gh.UploadURL, c.UploadURL.String())
	}
}
