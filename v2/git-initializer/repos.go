//go:build !testUnit && !lint
// +build !testUnit,!lint

// We exclude this file from unit tests and linting because it cannot be
// compiled without CGO and a specific version of libgit2 pre-installed. To keep
// our linting and unit tests lightweight, those are complications we'd like to
// avoid. We'll live without the linting and test this well with integration
// tests.

package main

import (
	"log"
	"strings"

	git "github.com/libgit2/git2go/v32"
	"github.com/pkg/errors"
)

// checkout checks out a specific commit, branch, or tag from the provided repo.
// Precedence is given to a specific commit, identified by the provided sha. If
// that value is empty, the provided reference will be used instead. If both are
// empty, the repository's default branch is checked out.
func checkout(repo *git.Repository, sha string, refStr string) error {
	// Our first concern is finding the commit in question and while we're at it
	// we'll create a local branch to check out into, if necessary.
	var commit *git.Commit
	var tagRef *git.Reference
	var localBranch *git.Branch
	var checkingOutDefaultBranch bool
	if sha != "" { // Specific commit identified by SHA takes precedence
		oid, err := git.NewOid(sha)
		if err != nil {
			return errors.Wrapf(err, "error getting oid from sha %q", sha)
		}
		if commit, err = repo.LookupCommit(oid); err != nil {
			return errors.Wrapf(err, "error getting commit %q", sha)
		}
		defer commit.Free()
		log.Printf("checking out commit %q", commit.Id().String())
	} else if refStr != "" { // Next in order of precedence is a specific ref
		ref, err := resolveRef(repo, refStr)
		if err != nil {
			return errors.Wrapf(err, "error resolving ref %q", refStr)
		}
		defer ref.Free()
		if commit, err = repo.LookupCommit(ref.Target()); err != nil {
			return errors.Wrapf(
				err,
				"error getting commit from ref %q",
				ref.Shorthand(),
			)
		}
		defer commit.Free()
		if ref.IsRemote() { // i.e. Not a tag
			headRef, err := repo.Head()
			if err != nil {
				return errors.Wrap(err, "error getting ref from default branch")
			}
			defer headRef.Free()
			if ref.Target().String() == headRef.Target().String() {
				// This ref points to the default branch, which we already have locally,
				// so we'll bypass creating a new local branch.
				checkingOutDefaultBranch = true
				log.Printf("checking out branch %q", headRef.Shorthand())
			} else { // We need to make a new local branch.
				// ref.Shorthand() will be like "origin/<branch or tag name>" and we
				// want JUST the branch or tag name.
				shortName := strings.SplitN(ref.Shorthand(), "/", 2)[1]
				if localBranch, err = repo.CreateBranch(shortName, commit, false); err != nil {
					return errors.Wrapf(
						err,
						"error creating branch %q from commit %q",
						shortName,
						commit.Id().String(),
					)
				}
				defer localBranch.Free()
				log.Printf("checking out branch %q", shortName)
			}
		} else {
			tagRef = ref
			log.Printf("checking out tag %q", ref.Shorthand())
		}
	} else { // Last in order of precedence is the default branch
		checkingOutDefaultBranch = true
		ref, err := repo.Head()
		if err != nil {
			return errors.Wrap(err, "error getting ref from default branch")
		}
		defer ref.Free()
		if commit, err = repo.LookupCommit(ref.Target()); err != nil {
			return errors.Wrap(
				err,
				"error getting commit from HEAD of default branch",
			)
		}
		log.Printf("checking out branch %q", ref.Shorthand())
	}

	// This is where we actually perform the checkout
	tree, err := repo.LookupTree(commit.TreeId())
	if err != nil {
		return errors.Wrapf(
			err,
			"error finding tree for commit %q",
			commit.Id().String(),
		)
	}
	defer tree.Free()
	if err = repo.CheckoutTree(
		tree,
		&git.CheckoutOptions{
			Strategy: git.CheckoutSafe |
				git.CheckoutRecreateMissing |
				git.CheckoutAllowConflicts |
				git.CheckoutUseTheirs,
		},
	); err != nil {
		return errors.Wrapf(
			err,
			"error checking out tree for commit %q",
			commit.Id().String(),
		)
	}

	// If we just checked out the default branch, we're done.
	if checkingOutDefaultBranch {
		return nil
	}

	// If we checked out some other branch, set HEAD to point to the local branch.
	if localBranch != nil {
		return errors.Wrapf(
			repo.SetHead(localBranch.Reference.Name()),
			"error setting HEAD to %q",
			localBranch.Reference.Name(),
		)
	}

	// If we checked out a tag, we want HEAD detached at the name of the tag.
	if tagRef != nil {
		return errors.Wrapf(
			repo.SetHead(tagRef.Name()),
			"error setting detached HEAD to %q",
			tagRef.Name(),
		)
	}

	// We must have checked out a specific commit by SHA. We want HEAD detached
	// at the SHA.
	return errors.Wrapf(
		repo.SetHeadDetached(commit.Id()),
		"error setting detached HEAD to %q",
		commit.Id().String(),
	)
}

// initSubmodules iterates over all of the provided repositories submodules and
// initializes and updates each.
func initSubmodules(repo *git.Repository) error {
	return repo.Submodules.Foreach(
		func(submodule *git.Submodule, name string) error {
			return errors.Wrapf(
				submodule.Update(true, nil),
				"error initializing submodule %q",
				name,
			)
		},
	)
}
