package main

import (
	"io/ioutil"
	"net/url"
	"path"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	gitssh "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

// setupAuth, if necessary, configures the git CLI for authentication using
// either SSH or the "store" (username/password-based) credential helper since
// the git-initializer component does fall back on the git CLI for certain
// operations. It additionally returns an appropriate implementation of
// transport.AuthMethod for operations that interact with remote repositories
// programmatically.
func setupAuth(evt event) (transport.AuthMethod, error) {
	homeDir, err := homedir.Dir()
	if err != nil {
		return nil, errors.Wrap(err, "error finding user's home directory")
	}

	// If an SSH key was provided, use that.
	if key, ok := evt.Project.Secrets["gitSSHKey"]; ok {
		// If a passphrase was supplied for the key, decrypt the key now.
		keyPass, ok := evt.Project.Secrets["gitSSHKeyPassword"]
		if ok {
			var err error
			if key, err = decryptKey(key, keyPass); err != nil {
				return nil, errors.Wrap(err, "error decrypting SSH key")
			}
		}
		rsaKeyPath := path.Join(homeDir, ".ssh", "id_rsa")
		if err := ioutil.WriteFile(rsaKeyPath, []byte(key), 0600); err != nil {
			return nil, errors.Wrapf(err, "error writing SSH key to %q", rsaKeyPath)
		}

		// This is the implementation of the transport.AuthMethod interface that can
		// be used for operations that interact with the remote repository
		// interactively.
		publicKeys, err := gitssh.NewPublicKeys("git", []byte(key), keyPass)
		if err != nil {
			return nil,
				errors.Wrap(err, "error getting transport.AuthMethod using SSH key")
		}

		// This prevents the CLI from interactively requesting the user to allow
		// connection to a new/unrecognized host.
		publicKeys.HostKeyCallback = ssh.InsecureIgnoreHostKey() // nolint: gosec
		return publicKeys, nil                                   // We're done
	}

	// If a password was provided, use that.
	if password, ok := evt.Project.Secrets["gitPassword"]; ok {
		credentialURL, err := url.Parse(evt.Worker.Git.CloneURL)
		if err != nil {
			return nil,
				errors.Wrapf(err, "error parsing URL %q", evt.Worker.Git.CloneURL)
		}
		// If a username was provided, use it. One may not have been because some
		// git providers, like GitHub, for instance, will allow any non-empty
		// username to be used in conjunction with a personal access token.
		username, ok := evt.Project.Secrets["gitUsername"]
		// If a username wasn't provided, we can ALSO try to pick it out of the URL.
		if !ok && credentialURL.User != nil {
			username = credentialURL.User.Username()
		}
		// If the username is still the empty string, we assume we're working with a
		// git provider like GitHub that only requires the username to be non-empty.
		// We arbitrarily set it to "git".
		if username == "" {
			username = "git"
		}
		// Remove path and query string components from the URL
		credentialURL.Path = ""
		credentialURL.RawQuery = ""
		// Augment the URL with user/pass information.
		credentialURL.User = url.UserPassword(username, password)
		// Write the URL to the location used by the "stored" credential helper.
		credentialsPath := path.Join(homeDir, ".git-credentials")
		if err := ioutil.WriteFile(
			credentialsPath,
			[]byte(credentialURL.String()),
			0600,
		); err != nil {
			return nil,
				errors.Wrapf(err, "error writing credentials to %q", credentialsPath)
		}

		// This is the implementation of the transport.AuthMethod interface that can
		// be used for operations that interact with the remote repository
		// interactively.
		return &http.BasicAuth{
			Username: username,
			Password: password,
		}, nil // We're done
	}

	// No auth setup required if we get to here.
	return nil, nil
}
