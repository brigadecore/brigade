package main

import (
	"fmt"
	"strings"

	git "github.com/libgit2/git2go/v32"
	"github.com/pkg/errors"
)

// resolveRef resolves refStr, which could be very specific and unambiguous
// (like "refs/heads/my-branch" -- which is clearly a branch) or relatively
// vague and ambiguous (like "foo" -- which could be a branch or a tag). This is
// the magic that makes the git-initializer DWIM (Do What I Mean).
func resolveRef(repo *git.Repository, refStr string) (*git.Reference, error) {
	if strings.HasPrefix(refStr, "refs/tags/") {
		// This refStr is very specific about what it is. Just look it up and return
		// whatever result we get.
		return repo.References.Lookup(refStr)
	}
	if strings.HasPrefix(refStr, "refs/heads/") {
		// This refStr is very specific about what it is. We need to factor the
		// remote name into the lookup, but then we can just return whatever result
		// we get.
		shortName := strings.SplitN(refStr, "/", 3)[2]
		return repo.References.Lookup(
			fmt.Sprintf("refs/remotes/origin/%s", shortName),
		)
	}
	// If we get to here, the refStr was a bit more vague about what it is.
	// Is it a tag?
	tagRefStr := fmt.Sprintf("refs/tags/%s", refStr)
	if ref, err := repo.References.Lookup(tagRefStr); err == nil {
		return ref, nil // No error. It WAS a tag!
	}
	// Is it a branch?
	branchRefStr := fmt.Sprintf("refs/remotes/origin/%s", refStr)
	if ref, err := repo.References.Lookup(branchRefStr); err == nil {
		return ref, nil // No error. It WAS a branch!
	}
	return nil,
		errors.Errorf("neither reference %q nor %q found", tagRefStr, branchRefStr)
}
