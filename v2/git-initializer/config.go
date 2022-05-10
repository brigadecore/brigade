package main

import (
	"os/exec"

	"github.com/pkg/errors"
)

// applyGlobalConfig applies global git CLI configuration options. Specifically,
// it adds /var/vcs as a "safe" directory. This is required due to the fact that
// the nonroot user the git-initializer process runs as doesn't OWN /var/vcs.
// (Kubernetes security context is applied to make /var/vcs group writable, but
// git can be pickier about file permissions than one would normally expect.)
func applyGlobalConfig() error {
	return errors.Wrapf(
		execCommand(
			exec.Command(
				"git",
				"config",
				"--global",
				"--add",
				"safe.directory",
				workspace,
			),
		),
		"error setting %q as a safe directory",
		workspace,
	)
}
