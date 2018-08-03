package webhook

import (
	"context"
	"testing"

	"github.com/google/go-github/github"

	"github.com/Azure/brigade/pkg/brigade"
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

type whClient struct {
	hook *github.Hook
}

func (w *whClient) CreateHook(context.Context, string, string, *github.Hook) (*github.Hook, *github.Response, error) {
	return nil, nil, nil
}

func (w *whClient) ListHooks(context.Context, string, string, *github.ListOptions) ([]*github.Hook, *github.Response, error) {
	if w.hook == nil {
		return nil, nil, nil
	}
	return []*github.Hook{w.hook}, nil, nil
}

func (w *whClient) DeleteHook(context.Context, string, string, int64) (*github.Response, error) {
	return nil, nil
}

func TestGHCreateHook(t *testing.T) {
	client := &whClient{}
	proj := &brigade.Project{}
	proj.Repo.Name = "github.com/Azure/brigade"
	callback := "http://localhost:7744"

	if err := createHook(client, proj, callback); err != nil {
		t.Error(err)
	}
}
