package main

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/brigadecore/brigade-foundations/retries"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/storage/filesystem"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/pkg/errors"
)

const (
	retryTimeLimit  = time.Minute
	retryMaxBackoff = 5 * time.Second
)

// cloneAndCheckoutCommit clones the remote repo specified by cloneURL and
// fetches and checks out only the tree identified by the provided SHA.
func cloneAndCheckoutCommit(cloneURL, sha string) error {
	// We have some repo configuration we want to handle before we attempt
	// cloning, so we start by initializing an empty repository.
	repo, err := git.Init(
		filesystem.NewStorage(
			osfs.New(filepath.Join(workspace, ".git")),
			cache.NewObjectLRUDefault(),
		),
		osfs.New(workspace),
	)
	if err != nil {
		return errors.Wrapf(err, "error initializing git repo at %q", workspace)
	}
	// Set the credential helper to "store" in case there's a username and
	// password in play. If there isn't this configuration causes no harm.
	// We fall back on the CLI to accomplish this because we'll also be falling
	// back on the CLI for the fetch.
	cmd := exec.Command(
		"git",
		"config",
		"credential.helper",
		"store",
	)
	cmd.Dir = workspace
	if err = execCommand(cmd); err != nil {
		return errors.Wrap(err, "error configuring credential helper")
	}

	// Define the remote
	if _, err = repo.CreateRemote(
		&config.RemoteConfig{
			Name: "origin",
			URLs: []string{cloneURL},
		},
	); err != nil {
		return errors.Wrapf(err, "error creating remote for %q", cloneURL)
	}

	// Fetch using the CLI. We use the CLI for this because the go-git library we
	// use for interacting with repositories programmatically does not implement
	// certain newer parts of the git spec that some providers -- namely Azure
	// DevOps -- require.
	ctx, cancel := context.WithTimeout(context.Background(), retryTimeLimit)
	defer cancel()
	if err = retries.ManageRetries(
		ctx, // Try until this context expires
		fmt.Sprintf("fetching %q", sha),
		0, // Infinite retries (until the context expires)
		retryMaxBackoff,
		func() (bool, error) {
			cmd = exec.Command(
				"git",
				"fetch",
				"origin",
				sha,
				"--no-tags",
			)
			cmd.Dir = workspace
			if err = execCommand(cmd); err != nil {
				err = errors.Wrapf(err, "error fetching %q", sha)
				// We only want to retry on DNS lookups because this is the error (on
				// Windows) that's indicative of the container's networking not being
				// ready yet.
				if strings.Contains(err.Error(), "Could not resolve host") {
					log.Println("Retrying...")
					return true, err
				}
				return false, err // Any other error, we'll surface right away
			}
			return false, nil // Success
		},
	); err != nil {
		return err
	}

	// The previous command should have pointed FETCH_HEAD at what we just
	// fetched, so check that out into the working tree.
	worktree, err := repo.Worktree()
	if err != nil {
		return errors.Wrap(err, "error getting repository's working tree")
	}
	return errors.Wrap(
		worktree.Checkout(
			&git.CheckoutOptions{
				Branch: "FETCH_HEAD",
			},
		),
		"error checking out FETCH_HEAD to working tree",
	)
}

// cloneAndCheckoutRef clones the remote repo specified by cloneURL and
// fetches and checks out only the tree identified by the symbolic reference.
func cloneAndCheckoutRef(cloneURL, ref string) error {
	ctx, cancel := context.WithTimeout(context.Background(), retryTimeLimit)
	defer cancel()
	// Here, a single branch clone executed via CLI works very nicely.
	return retries.ManageRetries(
		ctx, // Try until this context expires
		fmt.Sprintf("fetching %q", ref),
		0, // Infinite retries (until the context expires)
		retryMaxBackoff,
		func() (bool, error) {
			if err := execCommand(
				exec.Command(
					"git",
					"clone",
					"--branch",
					ref,
					"--single-branch",
					"--no-tags",
					"--config",
					"credential.helper=store",
					cloneURL,
					workspace,
				),
			); err != nil {
				err = errors.Wrapf(err, "error fetching %q", ref)
				// We only want to retry on DNS lookups because this is the error (on
				// Windows) that's indicative of the container's networking not being
				// ready yet.
				if strings.Contains(err.Error(), "Could not resolve host") {
					log.Println("Retrying...")
					return true, err
				}
				return false, err // Any other error, we'll surface right away
			}
			return false, nil // Success
		},
	)
}

// getDefaultBranch can find the default branch of a remote repository by
// listing all its references and looking for the one named HEAD. This is useful
// when we want to clone/checkout ONLY the default branch, but we don't already
// know its name.
func getDefaultBranch(
	cloneURL string,
	auth transport.AuthMethod,
) (string, error) {
	remote := git.NewRemote(
		memory.NewStorage(),
		&config.RemoteConfig{
			URLs: []string{cloneURL},
		},
	)
	var refs []*plumbing.Reference
	var err error
	ctx, cancel := context.WithTimeout(context.Background(), retryTimeLimit)
	defer cancel()
	if err = retries.ManageRetries(
		ctx, // Try until this context expires
		"listing refs",
		0, // Infinite retries (until the context expires)
		retryMaxBackoff,
		func() (bool, error) {
			if refs, err = remote.List(
				&git.ListOptions{
					Auth: auth,
				},
			); err != nil {
				err = errors.Wrap(err, "error listing refs")
				// We only want to retry on DNS lookups because this is the error (on
				// Windows) that's indicative of the container's networking not being
				// ready yet.
				if strings.Contains(err.Error(), "Could not resolve host") {
					log.Println("Retrying...")
					return true, err
				}
				return false, err // Any other error, we'll surface right away
			}
			return false, nil // Success
		},
	); err != nil {
		return "", err
	}
	// Now look through the refs...
	for _, ref := range refs {
		if ref.Name() == plumbing.HEAD {
			return ref.Target().Short(), nil
		}
	}
	return "", errors.New("failed to find default branch")
}

// getShortRef recognizes refs of the form refs/tags/* and refs/heads/* and
// returns only the "short" name of the ref -- which is the input expected by
// the `--branch`` flag on a `git clone`.
func getShortRef(ref string) string {
	if strings.HasPrefix(ref, "refs/tags/") ||
		strings.HasPrefix(ref, "refs/heads/") {
		return strings.SplitAfterN(ref, "/", 3)[2]
	}
	return ref
}

// initSubmodules initializes and updates submodules.
func initSubmodules() error {
	cmd := exec.Command("git", "submodule", "update", "--init", "--recursive")
	cmd.Dir = workspace
	return errors.Wrap(execCommand(cmd), "error initializing submodules")
}
