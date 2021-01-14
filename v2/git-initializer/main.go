package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

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

	"github.com/brigadecore/brigade/sdk/v2/core"
)

func main() {
	if err := gitCheckout(); err != nil {
		fmt.Printf("\n%s\n\n", err)
		os.Exit(1)
	}
}

// TODO: needs retry - let's do a follow-up (krancour has a lib to use!)

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
	// TODO: What about token-based auth?  (see v1 askpass.sh/BRIGADE_REPO_AUTH_TOKEN)

	// TODO: Check for SSH Cert
	// https://github.blog/2019-08-14-ssh-certificate-authentication-for-github-enterprise-cloud/
	//
	// (BRIGADE_REPO_SSH_CERT in v1)
	// I think all we need to do is to make sure the cert file exists at /id_dsa-cert.pub

	// Check for SSH Key
	privateKey, ok := event.Project.Secrets["gitSSHKey"]
	if ok {
		publicKeys, err := gitssh.NewPublicKeys(
			"git",
			[]byte(privateKey),
			event.Project.Secrets["gitSSHKeyPassword"],
		)
		if err != nil {
			return errors.Wrapf(
				err,
				"error configuring authentication for remote with URL %s: %s",
				event.Worker.Git.CloneURL,
			)
		}
		// The following avoids:
		// "unable to find any valid known_hosts file, set SSH_KNOWN_HOSTS env variable"
		publicKeys.HostKeyCallback = ssh.InsecureIgnoreHostKey()
		auth = publicKeys
	}

	// Prioritize using Commit; alternatively try Ref
	commitRef := event.Worker.Git.Commit
	if commitRef == "" {
		commitRef = event.Worker.Git.Ref // This will never be non-empty
	}
	fullRef := plumbing.NewReferenceFromStrings(commitRef, commitRef)
	refSpec := config.RefSpec(fmt.Sprintf("+%s:%s", fullRef.Name(), fullRef.Name()))

	// Initialize an empty repository with an empty working tree
	workspace := "/var/vcs"
	gitStorage := filesystem.NewStorage(
		osfs.New(filepath.Join(workspace, ".git")),
		cache.NewObjectLRUDefault(),
	)

	repo, err := git.Init(gitStorage, osfs.New(workspace))
	if err != nil {
		return errors.Wrapf(err, "error initializing git repository at %s: %s", workspace)
	}

	const remoteName = "origin"
	// If we're not dealing with an exact commit, list the remote refs
	// to build out a full, updated refspec
	//
	// For example, we might be supplied with the tag "v0.1.0", but if we use
	// this directly, we'll get: couldn't find remote ref "v0.1.0"
	// So we need to find the full remote ref; in this case, "refs/tags/v0.1.0"
	if gitConfig.Commit == "" {
		// Create a new remote for the purposes of listing remote refs and finding
		// the full ref we want
		remoteConfig := &config.RemoteConfig{
			Name:  remoteName,
			URLs:  []string{gitConfig.CloneURL},
			Fetch: []config.RefSpec{refSpec},
		}
		rem := git.NewRemote(gitStorage, remoteConfig)

		// List remote refs
		refs, err := rem.List(&git.ListOptions{Auth: auth})
		if err != nil {
			return errors.Wrap(err, "error listing remotes")
		}

		// Filter the list of refs and only keep the full ref matching our commitRef
		var found bool
		for _, ref := range refs {
			// Ignore the HEAD symbolic reference
			// e.g. [HEAD ref: refs/heads/master]
			if ref.Type() == plumbing.SymbolicReference {
				continue
			}

			// TODO: update this, as it is faulty... there may be multiple matches
			// e.g. "main" might match /refs/heads/main and /refs/heads/main2
			if strings.Contains(ref.Name().String(), fullRef.Name().String()) ||
				strings.Contains(ref.Hash().String(), fullRef.Hash().String()) {
				fullRef = ref
				found = true
			}
		}

		if !found {
			return fmt.Errorf("reference %q not found in repo %q", fullRef.Name(), gitConfig.CloneURL)
		}

		// Create refspec with the updated ref
		refSpec = config.RefSpec(fmt.Sprintf("+%s:%s",
			fullRef.Name(), fullRef.Name()))
	}

	// Create the remote that we'll use to fetch, using the updated/full refspec
	remoteConfig := &config.RemoteConfig{
		Name:  remoteName,
		URLs:  []string{gitConfig.CloneURL},
		Fetch: []config.RefSpec{refSpec},
	}
	remote, err := repo.CreateRemote(remoteConfig)
	if err != nil {
		return errors.Wrap(err, "error creating remote")
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
		return errors.Wrap(err, "unable to set ref")
	}

	// Fetch the ref specs we are interested in
	fetchOpts := &git.FetchOptions{
		RemoteName: remoteName,
		RefSpecs:   []config.RefSpec{refSpec},
		Force:      true,
		Auth:       auth,
		Progress:   os.Stdout,
	}
	err = remote.Fetch(fetchOpts)
	if err != nil {
		return errors.Wrap(err, "error fetching refs from the remote")
	}

	// Get the repository's working tree
	worktree, err := repo.Worktree()
	if err != nil {
		return errors.Wrap(err, "unable to access the repo worktree")
	}

	// Check out whatever we're interested in into the working tree
	if err = worktree.Checkout(
		&git.CheckoutOptions{
			Branch: fullRef.Name(),
			Force:  true,
		},
	); err != nil {
		return errors.Wrapf(err, "unable to checkout using %q", commitRef)
	}

	// Initialize submodules if configured to do so
	if event.Worker.Git.InitSubmodules {
		submodules, err := worktree.Submodules()
		if err != nil {
			return errors.Wrap(err, "error retrieving submodules: %s")
		}
		for _, submodule := range submodules {
			if err = submodule.Update(
				&git.SubmoduleUpdateOptions{
					Init:              true,
					RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
				},
			); err != nil {
				return errors.Wrapf(
					err,
					"error updating submodule %q: %s",
					submodule.Config().Name,
				)
			}
		}
	}

	return nil
}
