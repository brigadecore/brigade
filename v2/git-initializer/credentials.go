package main

import (
	git "github.com/libgit2/git2go/v32"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

// getCredentialsCallback extracts credentials, if any, from project secrets in
// the provided map and returns a callback function. If no credentials are found
// in the secrets, a nil function pointer is returned.
func getCredentialsCallback(
	secrets map[string]string,
) (git.CredentialsCallback, error) {
	// We'll check the project's secrets first for a well-known key that we could
	// expect contains a private SSH key. If we find that, we'll use it.
	if privateKey, ok := secrets["gitSSHKey"]; ok {
		var signer ssh.Signer
		var err error
		// The private key may or may not be protected by a passphrase...
		if keyPassStr, ok := secrets["gitSSHKeyPassword"]; ok {
			// Auth using key and passphrase
			signer, err = ssh.ParsePrivateKeyWithPassphrase(
				[]byte(privateKey),
				[]byte(keyPassStr),
			)
		} else {
			// Auth using a key without a passphrase
			signer, err = ssh.ParsePrivateKey([]byte(privateKey))
		}
		if err != nil {
			return nil, errors.Wrap(
				err,
				`error parsing private SSH key specified by secret "gitSSHKey"`,
			)
		}
		return func(
			string,
			string,
			git.CredentialType,
		) (*git.Credential, error) {
			return git.NewCredentialSSHKeyFromSigner("git", signer)
		}, nil
	}

	// Check the project's secrets for a well-known key that we could expect
	// contains a password or token.
	if password, ok := secrets["gitPassword"]; ok {
		// There may or may not be a username associated with the password. It
		// really depends on who hosts the repository we're cloning from. GitHub,
		// for instance, expects the username to be any non-empty string (and the
		// password to be a personal access token), while Bitbucket expects a valid
		// username (and the password to be an app password).
		username := secrets["gitUsername"]
		// Ultimately, the username and password are used with basic auth and an
		// empty username isn't allowed, so if no username was specified (e.g. for a
		// repo hosted on GitHub), just use the string "git" as the username.
		if username == "" {
			username = "git"
		}
		return func(
			string,
			string,
			git.CredentialType,
		) (*git.Credential, error) {
			// Basic auth
			return git.NewCredentialUserpassPlaintext(username, password)
		}, nil
	}

	// No credentials found
	return nil, nil
}
