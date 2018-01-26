package commands

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/api/core/v1"

	"github.com/Azure/brigade/brigade-vacuum/cmd/brigade-vacuum/vacuum"
	"github.com/Azure/brigade/pkg/storage/kube"
)

const (
	envKubeConfig = "KUBECONFIG"
	envMaxBuilds  = "VACUUM_MAX_BUILDS"
	envAge        = "VACUUM_AGE"
	envNamespace  = "BRIGADE_NAMESPACE"
)

const mainUsage = `Clean up old Brigade builds

brigade-vacuum deletes old data left by builds. It is configurable so that
administrators can determine how often they want to flush data from the system.

There are two ways to identify which records should be removed:

- age: Any records older than the given age will be vacuumed up.
- max-builds: An absolute limit on the number of builds retained.

The two can be combined. Setting a zero-value for either one will remove the limit.

AGE VALUES
==========

The '--age' parameter can take a valid string duration with any of these suffixes:

- h: hour
- m: minute
- s: second
- ns, us, and ms: nano-, µ-, and milli-seconds.

Note that unless the vacuum is run with high frequency, smaller values are not
particularly useful.
`

var (
	globalKubeConfig = ""
	globalNamespace  = ""
	globalAge        = ""
	globalVerbose    = false
	globalMaxBuilds  = -1
)

func init() {
	f := Root.PersistentFlags()
	f.StringVarP(&globalNamespace, "namespace", "n", "", "The Kubernetes namespace for Brigade")
	f.StringVarP(&globalAge, "age", "a", "", "Age as a fuzzy date ('48h' for hours, '20m' for minutes, '2000s' for seconds)")
	f.IntVarP(&globalMaxBuilds, "max-builds", "m", 0, "Maxinum number of builds to keep")
	f.BoolVarP(&globalVerbose, "verbose", "v", false, "Turn on verbose output")
	f.StringVar(&globalKubeConfig, "kubeconfig", "", "The path to a KUBECONFIG file, overrides $KUBECONFIG.")
}

// Root is the top-level command, which just prints help text.
var Root = &cobra.Command{
	Use:   "brigade-vacuum",
	Short: "Clean up old Brigade builds",
	Long:  mainUsage,
	RunE: func(cmd *cobra.Command, args []string) error {
		a := getAge()
		mb := maxBuilds()
		if a == "" && mb == 0 {
			return errors.New("one of --age or --max-builds must be greater than zero")
		}
		var age = time.Time{}
		if a != "" {
			dur, err := time.ParseDuration(a)
			if err != nil {
				return err
			}
			age = time.Now().Add(-dur)
		}
		c, err := kube.GetClient("", kubeConfigPath())
		if err != nil {
			return err
		}
		if globalVerbose {
			fmt.Fprintf(os.Stderr, "Max Age: %s\nMax Builds: %d\n", age, globalMaxBuilds)
		}
		count, err := vacuum.New(age, mb, c, ns()).Run()
		fmt.Fprintf(os.Stdout, "Deleted %d\n", count)
		return err
	},
}

func kubeConfigPath() string {
	if globalKubeConfig != "" {
		return globalKubeConfig
	}
	if v, ok := os.LookupEnv(envKubeConfig); ok {
		return v
	}
	defConfig := os.ExpandEnv("$HOME/.kube/config")
	if _, err := os.Stat(defConfig); err == nil {
		fmt.Printf("Using config from %s\n", defConfig)
		return defConfig
	}

	// If we get here, we might be in-Pod.
	return ""
}

func ns() string {
	if globalNamespace != "" {
		return globalNamespace
	}
	if v, ok := os.LookupEnv(envNamespace); ok {
		return v
	}
	return v1.NamespaceDefault
}

func maxBuilds() int {
	if globalMaxBuilds > -1 {
		return globalMaxBuilds
	}
	v, ok := os.LookupEnv(envMaxBuilds)
	if !ok {
		return 0
	}
	m, err := strconv.Atoi(v)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ignoring non-int value %q", v)
		return 0
	}
	return m
}

func getAge() string {
	if globalAge != "" {
		return globalAge
	}
	return os.Getenv(envAge)
}
