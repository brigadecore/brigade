package main

import (
	"fmt"
	"os"

	"github.com/Masterminds/vcs"
)

var envDefaults = map[string]string{
	"VCS_REPO":       "",
	"VCS_REVISION":   "master",
	"VCS_LOCAL_PATH": "/home",
}

func main() {
	env := make(map[string]string, len(envDefaults))
	for key, val := range envDefaults {
		env[key] = envOr(key, val)
	}

	err := checkout(env["VCS_REPO"], env["VCS_LOCAL_PATH"], env["VCS_REVISION"])
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAILED: %s\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stdout, "Successfully fetched %s", env["VCS_REPO"])
}

func envOr(varname, defaultVal string) string {
	if val, ok := os.LookupEnv(varname); ok {
		return val
	}
	return defaultVal
}

func checkout(repo, dest, rev string) error {
	r, err := vcs.NewRepo(repo, dest)
	if err != nil {
		// Should this be a warning?
		return err
	}
	if err := r.Get(); err != nil {
		// We warn here because an init container can restart with the
		// same pod.
		fmt.Fprintf(os.Stdout, "WARNING: %s", err)
	}
	return r.UpdateVersion(rev)
}
