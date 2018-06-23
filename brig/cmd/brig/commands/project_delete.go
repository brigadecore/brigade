package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/spf13/cobra"

	"github.com/Azure/brigade/pkg/storage/kube"
)

var projectDeleteUsage = `Delete a project.

This removes the project from brigade by deleting the project secret in Kubernetes.

It does not delete old builds or jobs.

With the -o/--out flag, you can save a local backup copy of this project before deleting
it.
`

var (
	projectDeleteOut    = ""
	projectDeleteDryRun = false
)

func init() {
	project.AddCommand(projectDelete)
	flags := projectDelete.Flags()
	flags.StringVarP(&projectDeleteOut, "out", "o", "", "File where configuration should be saved before deletion")
	flags.BoolVarP(&projectDeleteDryRun, "dry-run", "D", false, "Check that the project exists, but don't delete it.")
}

var projectDelete = &cobra.Command{
	Use:   "delete PROJECT",
	Short: "delete a Brigade project",
	Long:  projectDeleteUsage,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("project name is a required argument")
		}
		return deleteProject(cmd.OutOrStderr(), args[0])
	},
}

func deleteProject(out io.Writer, pid string) error {
	c, err := kubeClient()
	if err != nil {
		return err
	}

	store := kube.New(c, globalNamespace)
	p, err := store.GetProject(pid)
	if err != nil {
		// To remain idempotent, we just return nil here.
		if globalVerbose {
			fmt.Fprintf(out, "project not found: %q\n", pid)
		}
		return nil
	}

	if projectDeleteOut != "" {
		secret, err := kube.SecretFromProject(p)
		if err != nil {
			return fmt.Errorf("deletion aborted due to conversion error: %s", err)
		}
		data, err := json.Marshal(secret)
		if err != nil {
			return fmt.Errorf("deletion aborted due to encoding error: %s", err)
		}
		if err := ioutil.WriteFile(projectDeleteOut, data, 0755); err != nil {
			return fmt.Errorf("deletion aborted because backup could not be created: %s", err)
		}
	}

	if globalVerbose {
		fmt.Fprintf(out, "Deleting %s\n", p.ID)
	}
	return store.DeleteProject(p.ID)
}
