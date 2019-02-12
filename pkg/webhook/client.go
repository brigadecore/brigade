package webhook

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"

	"github.com/Azure/brigade/pkg/brigade"
)

// ghClient gets a new GitHub client object.
//
// It authenticates with an OAUTH2 token.
//
// If an enterpriseHost base URL is provided, this will attempt to connect to
// that instead of the hosted GitHub API server.
func ghClient(gh brigade.Github) (*github.Client, error) {
	t := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: gh.Token})
	c := context.Background()
	tc := oauth2.NewClient(c, t)
	if gh.BaseURL != "" {
		return github.NewEnterpriseClient(gh.BaseURL, gh.UploadURL, tc)
	}
	return github.NewClient(tc), nil
}

// GetFileContents returns the contents for a particular file in the project.
func GetFileContents(proj *brigade.Project, ref, path string) ([]byte, error) {
	c := context.Background()

	var client *github.Client
	if proj.Github.Token != "" { // GitHub project configured with Auth Token
		var err error
		client, err = ghClient(proj.Github)
		if err != nil {
			return []byte{}, err
		}
	} else { // OSS project
		netClient := &http.Client{ // https://medium.com/@nate510/don-t-use-go-s-default-http-client-4804cb19f779
			Timeout: time.Second * 30, // 30 seconds timeout should be enough
		}
		client = github.NewClient(netClient)
	}

	parts := strings.SplitN(proj.Repo.Name, "/", 3)
	if len(parts) != 3 {
		return nil, fmt.Errorf("project name %q is malformed", proj.Repo.Name)
	}
	opts := &github.RepositoryContentGetOptions{Ref: ref}

	r, err := client.Repositories.DownloadContents(c, parts[1], parts[2], path, opts)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return ioutil.ReadAll(r)

}
