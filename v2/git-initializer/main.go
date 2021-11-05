package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/plumbing/transport"
	gitssh "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/filesystem"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"

	"github.com/brigadecore/brigade-foundations/retries"
	"github.com/brigadecore/brigade-foundations/version"
	"github.com/brigadecore/brigade/sdk/v2/core"
)

const (
	workspace     = "/var/vcs"
	maxRetryCount = 5
	maxBackoff    = 5 * time.Second
)

func main() {
	log.Printf(
		"Starting Brigade Git Initializer -- version %s -- commit %s",
		version.Version(),
		version.Commit(),
	)

	if err := gitCheckout(); err != nil {
		fmt.Printf("\n%s\n\n", err)
		os.Exit(1)
	}
}

// nolint: gocyclo
func gitCheckout() error {
	eventPath := "/var/event/event.json"
	data, err := ioutil.ReadFile(eventPath)
	if err != nil {
		return errors.Wrapf(err, "unable read the event file %q", eventPath)
	}

	var event struct {
		Project struct {
			Secrets map[string]string `json:"secrets"`
		} `json:"project"`
		Worker struct {
			Git *core.GitConfig `json:"git"`
		} `json:"worker"`
	}
	err = json.Unmarshal(data, &event)
	if err != nil {
		return errors.Wrap(err, "error unmarshaling the event")
	}

	// Extract git config
	gitConfig := event.Worker.Git
	if gitConfig == nil {
		return fmt.Errorf("git config from %q is empty", eventPath)
	}

	// Setup Auth
	var auth transport.AuthMethod
	// TODO: What about token-based auth?
	// (see v1 askpass.sh/BRIGADE_REPO_AUTH_TOKEN)

	// TODO: Check for SSH Cert
	// (see https://github.com/brigadecore/brigade/pull/1008)

	// Check for SSH Key
	privateKey, ok := event.Project.Secrets["gitSSHKey"]
	if ok {
		var publicKeys *gitssh.PublicKeys
		publicKeys, err = gitssh.NewPublicKeys(
			"git",
			[]byte(privateKey),
			event.Project.Secrets["gitSSHKeyPassword"],
		)
		if err != nil {
			return errors.Wrapf(
				err,
				"error configuring authentication for remote with URL %s",
				event.Worker.Git.CloneURL,
			)
		}
		// The following avoids:
		// "unable to find any valid known_hosts file,
		// set SSH_KNOWN_HOSTS env variable"
		publicKeys.HostKeyCallback = ssh.InsecureIgnoreHostKey() // nolint: gosec
		auth = publicKeys
	}

	var worktree *git.Worktree
	if event.Worker.Git.Commit == "" && event.Worker.Git.Ref == "" {
		worktree, err = simpleCheckout(gitConfig.CloneURL, auth)
	} else {
		worktree, err = complexCheckout(
			gitConfig.CloneURL,
			event.Worker.Git.Commit,
			event.Worker.Git.Ref,
			auth,
		)
	}
	if err != nil {
		return err
	}

	// Initialize submodules if configured to do so
	if event.Worker.Git.InitSubmodules {
		err = initSubmodules(worktree)
	}
	return err
}

func simpleCheckout(
	cloneURL string,
	auth transport.AuthMethod,
) (*git.Worktree, error) {
	// This will reliably get whichever branch is designated as the default.
	repo, err := git.PlainClone(
		workspace,
		false,
		&git.CloneOptions{
			URL:      cloneURL,
			Auth:     auth,
			Progress: os.Stdout,
		},
	)
	if err != nil {
		return nil, errors.Wrapf(err, "error cloning repository at %s", cloneURL)
	}
	worktree, err := repo.Worktree()
	return worktree, errors.Wrapf(
		err,
		"error getting worktree from repository cloned from %s",
		cloneURL,
	)
}

func complexCheckout(
	cloneURL string,
	commit string,
	ref string,
	auth transport.AuthMethod,
) (*git.Worktree, error) {
	// Prioritize using Commit; alternatively try Ref
	commitRef := commit
	if commitRef == "" {
		commitRef = ref // This will never be empty
	}
	fullRef := plumbing.NewReferenceFromStrings(commitRef, commitRef)
	refSpec := config.RefSpec(
		fmt.Sprintf("+%s:%s", fullRef.Name(), fullRef.Name()))

	// Initialize an empty repository with an empty working tree
	gitStorage := filesystem.NewStorage(
		osfs.New(filepath.Join(workspace, ".git")),
		cache.NewObjectLRUDefault(),
	)

	repo, err := git.Init(gitStorage, osfs.New(workspace))
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"error initializing git repository at %s",
			workspace,
		)
	}

	const remoteName = "origin"
	// If we're not dealing with an exact commit, and we don't already have a
	// full reference, list the remote refs to build out a full, updated refspec
	//
	// For example, we might be supplied with the tag "v0.1.0", but if we use
	// this directly, we'll get: couldn't find remote ref "v0.1.0"
	// So we need to find the full remote ref; in this case, "refs/tags/v0.1.0"
	if commit == "" && !isFullReference(fullRef.Name()) {
		// Create a new remote for the purposes of listing remote refs and finding
		// the full ref we want
		remoteConfig := &config.RemoteConfig{
			Name:  remoteName,
			URLs:  []string{cloneURL},
			Fetch: []config.RefSpec{refSpec},
		}
		rem := git.NewRemote(gitStorage, remoteConfig)

		// List remote refs
		var refs []*plumbing.Reference
		refs, err = rem.List(&git.ListOptions{Auth: auth})
		if err != nil {
			return nil, errors.Wrap(err, "error listing remotes")
		}

		// Filter the list of refs and only keep the full ref matching our commitRef
		matches := make([]*plumbing.Reference, 0)
		for _, ref := range refs {
			// Ignore the HEAD symbolic reference
			// e.g. [HEAD ref: refs/heads/main]
			if ref.Type() == plumbing.SymbolicReference {
				continue
			}

			// Compare the short names of both refs,
			// where the short name of e.g. '/refs/heads/main' is 'main'
			// Alternatively, match on ref hash
			if ref.Name().Short() == fullRef.Name().Short() ||
				ref.Hash() == fullRef.Hash() {
				matches = append(matches, ref)
			}
		}

		if len(matches) == 0 {
			return nil, fmt.Errorf(
				"reference %q not found in repo %q",
				fullRef.Name(),
				cloneURL,
			)
		}
		if len(matches) > 1 {
			return nil, fmt.Errorf(
				"found more than one match for reference %q: %+v",
				fullRef.Name(),
				matches,
			)
		}
		fullRef = matches[0]

		// Create refspec with the updated ref
		refSpec = config.RefSpec(fmt.Sprintf("+%s:%s",
			fullRef.Name(), fullRef.Name()))
	}

	// Create the remote that we'll use to fetch, using the updated/full refspec
	remoteConfig := &config.RemoteConfig{
		Name:  remoteName,
		URLs:  []string{cloneURL},
		Fetch: []config.RefSpec{refSpec},
	}
	remote, err := repo.CreateRemote(remoteConfig)
	if err != nil {
		return nil, errors.Wrap(err, "error creating remote")
	}

	// Create a FETCH_HEAD reference pointing to our ref hash
	// The go-git library doesn't appear to support adding this, though the
	// git CLI does.  This is for parity with v1 functionality.
	//
	// From https://git-scm.com/docs/gitrevisions:
	// "FETCH_HEAD records the branch which you fetched from a remote repository
	// with your last git fetch invocation."
	newRef := plumbing.NewReferenceFromStrings("FETCH_HEAD",
		fullRef.Hash().String())
	err = repo.Storer.SetReference(newRef)
	if err != nil {
		return nil, errors.Wrap(err, "unable to set ref")
	}

	// Fetch the ref specs we are interested in
	fetchOpts := &git.FetchOptions{
		RemoteName: remoteName,
		RefSpecs:   []config.RefSpec{refSpec},
		Force:      true,
		Auth:       auth,
		Progress:   os.Stdout,
	}
	if retryErr := retries.ManageRetries(
		context.Background(),
		"git fetch",
		maxRetryCount,
		maxBackoff,
		func() (bool, error) {
			err = remote.Fetch(fetchOpts)
			if err != nil {
				return true, errors.Wrap(err, "error fetching refs from the remote")
			}
			return false, nil
		},
	); retryErr != nil {
		return nil, retryErr
	}

	// Get the repository's working tree
	worktree, err := repo.Worktree()
	if err != nil {
		return nil, errors.Wrap(err, "unable to access the repo worktree")
	}

	// Check out whatever we're interested in into the working tree
	if retryErr := retries.ManageRetries(
		context.Background(),
		"git checkout",
		maxRetryCount,
		maxBackoff,
		func() (bool, error) {
			err := worktree.Checkout(
				&git.CheckoutOptions{
					Branch: fullRef.Name(),
					Force:  true,
				},
			)
			if err != nil {
				return true, errors.Wrapf(
					err,
					"unable to checkout using %q",
					commitRef,
				)
			}
			return false, nil
		},
	); retryErr != nil {
		return nil, retryErr
	}

	return worktree, nil
}

func initSubmodules(worktree *git.Worktree) error {
	submodules, err := worktree.Submodules()
	if err != nil {
		return errors.Wrap(err, "error retrieving submodules: %s")
	}
	if retryErr := retries.ManageRetries(
		context.Background(),
		"update submodules",
		maxRetryCount,
		maxBackoff,
		func() (bool, error) {
			for _, submodule := range submodules {
				if err = submodule.Update(
					&git.SubmoduleUpdateOptions{
						Init:              true,
						RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
					},
				); err != nil {
					return true, errors.Wrapf(
						err,
						"error updating submodule %q",
						submodule.Config().Name,
					)
				}
			}
			return false, nil
		},
	); retryErr != nil {
		return retryErr
	}
	return nil
}

func isFullReference(name plumbing.ReferenceName) bool {
	return name.IsBranch() || name.IsNote() || name.IsRemote() || name.IsTag()
}
