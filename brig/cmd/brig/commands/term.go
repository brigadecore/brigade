package commands

import (
	"time"

	"github.com/rivo/tview"
	"github.com/slok/brigadeterm/pkg/controller"
	"github.com/slok/brigadeterm/pkg/service/brigade"
	"github.com/slok/brigadeterm/pkg/ui"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"

	"github.com/brigadecore/brigade/pkg/storage/kube"
)

var reloadInterval string

func init() {
	Root.AddCommand(term)

	flags := term.PersistentFlags()
	flags.StringVarP(&reloadInterval, "reload-interval", "", "3s", "The interval the UI will autoreload")

}

var term = &cobra.Command{
	Use:   "term",
	Short: "Starts a terminal dashboard",
	Long:  checkUsage,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTerm()
	},
}

func runTerm() error {
	config, err := getKubeConfig()
	if err != nil {
		return err
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	store := kube.New(client, globalNamespace)
	svc := brigade.NewService(store)
	uictrl := controller.NewController(svc)
	app := tview.NewApplication()
	t, err := time.ParseDuration(reloadInterval)
	if err != nil {
		return err
	}
	index := ui.NewIndex(t, uictrl, app)
	return index.Render()
}
