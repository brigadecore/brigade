package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"time"

	"k8s.io/client-go/kubernetes"

	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"

	"k8s.io/apimachinery/pkg/util/duration"

	"github.com/brigadecore/brigade/pkg/brigade"
	"github.com/brigadecore/brigade/pkg/storage/kube"
)

const buildListUsage = `List all installed builds.

Print a list of the current builds starting from latest (in creation time) to oldest. By default it will print all the builds, use --count to get a subset of them.
`

var (
	buildListCount int
	output         string
)

func init() {
	build.AddCommand(buildList)
	buildList.Flags().IntVarP(&buildListCount, "count", "c", 0, "The maximum number of builds to return. 0 for all")
	buildList.Flags().StringVarP(&output, "output", "o", "", "Return output in another format. Supported formats: json")
}

var buildList = &cobra.Command{
	Use:   "list [project]",
	Short: "list builds",
	Long:  buildListUsage,
	RunE: func(cmd *cobra.Command, args []string) error {
		proj := ""
		if len(args) > 0 {
			proj = args[0]
		}

		c, err := kubeClient()
		if err != nil {
			return err
		}

		bls, err := getBuilds(proj, c, buildListCount)
		if err != nil {
			return err
		}

		if output == "json" {
			bj, err := json.MarshalIndent(bls, "", "    ")
			if err != nil {
				return err
			}
			_, err = cmd.OutOrStdout().Write(bj)
			return err
		}

		listBuilds(getBuildsForStdout(bls), cmd.OutOrStdout())
		return nil
	},
}

func listBuilds(bs []*buildForStdout, out io.Writer) {
	table := uitable.New()
	table.AddRow("ID", "TYPE", "PROVIDER", "PROJECT", "STATUS", "AGE")
	for _, b := range bs {
		table.AddRow(b.ID, b.Type, b.Provider, b.ProjectID, b.status, b.since)
	}
	fmt.Fprintln(out, table)
}

func getBuilds(project string, client kubernetes.Interface, count int) ([]*brigade.Build, error) {
	var builds []*brigade.Build
	var err error

	store := kube.New(client, globalNamespace)
	if project == "" {
		builds, err = store.GetBuilds()
		if err != nil {
			return nil, err
		}
	} else {
		proj, err := store.GetProject(project)
		if err != nil {
			return nil, err
		}
		builds, err = store.GetProjectBuilds(proj)
		if err != nil {
			return nil, err
		}
	}

	// sorting here on StartTime because we do not want to rely on K8s sorting
	// which would be the order that Secrets/Pods were created
	sort.Slice(builds, func(i, j int) bool {
		if builds[i].Worker == nil || builds[j].Worker == nil {
			return false
		}
		return builds[i].Worker.StartTime.After(builds[j].Worker.StartTime)
	})

	if count == 0 || count > len(builds) {
		count = len(builds)
	}

	return builds[:count], err
}

func getBuildsForStdout(builds []*brigade.Build) []*buildForStdout {
	var bfss []*buildForStdout
	for i := 0; i < len(builds); i++ {

		b := builds[i]
		bfs := &buildForStdout{Build: builds[i]}

		bfs.status = "???"
		bfs.since = "???"
		if b.Worker != nil {
			bfs.status = b.Worker.Status.String()
			if b.Worker.Status == brigade.JobSucceeded || b.Worker.Status == brigade.JobFailed {
				bfs.since = duration.ShortHumanDuration(time.Since(b.Worker.StartTime))
			}
		}
		bfss = append(bfss, bfs)
	}

	return bfss
}

type buildForStdout struct {
	*brigade.Build
	status string
	since  string
}
