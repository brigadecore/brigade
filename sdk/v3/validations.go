package sdk

import (
	"fmt"
	"regexp"
)

var projectIDRegex = regexp.MustCompile(`^[a-z][a-z\d-]*[a-z\d]$`)

// nolint: lll
var gitCloneURLRegex = regexp.MustCompile(`^(?:(?:https?://)|(?:git@))[\w:/\-\.\?=@&%]+$`)

// ValidateProjectID checks if a given id is valid.
func ValidateProjectID(id string) error {
	idMatch := projectIDRegex.MatchString(id)
	if !idMatch || len(id) < 3 || len(id) > 63 {
		return fmt.Errorf(
			"invalid value %q for project id"+
				" (3-63 characters, first char must be"+
				" a letter, lowercase letters only)",
			id,
		)
	}
	return nil
}

// ValidateProjectID checks if a given git clone URL is valid by ensuring it
// begins with "http://", "https://", or "git@".
func ValidateGitCloneURL(url string) error {

	urlMatch := gitCloneURLRegex.MatchString(url)
	if !urlMatch {
		return fmt.Errorf(
			"invalid value %q for git clone URL"+
				" (must start with http://, https://, or git@)",
			url,
		)
	}
	return nil
}
