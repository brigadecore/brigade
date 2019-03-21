package commands

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bacongobbler/browser"
	"github.com/spf13/cobra"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"

	"github.com/brigadecore/brigade/pkg/portforwarder"
)

const (
	dashboardUsage = `Opens the Kashti dashboard with the default browser.
`
	remotePort = 80
)

var (
	port          int
	openDashboard bool
)

func init() {
	Root.AddCommand(dashboard)

	flags := dashboard.PersistentFlags()
	flags.IntVar(&port, "port", 8081, "local port for the Kashti dashboard")
	flags.BoolVar(&openDashboard, "open-dashboard", true, "open the dashboard in the browser")
}

var dashboard = &cobra.Command{
	Use:   "dashboard",
	Short: "Open the Kashti dashboard",
	Long:  dashboardUsage,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Printf("Connecting to kashti at http://localhost:%d...\n", port)

		tunnel, err := startProxy(port)
		if err != nil {
			return err
		}
		defer tunnel.Close()

		stop := make(chan os.Signal, 2)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-stop
			os.Exit(0)
		}()

		log.Println("Connected! When you are finished with this session, enter CTRL+C.")

		if openDashboard {
			if err := browser.Open(fmt.Sprintf("http://localhost:%d", port)); err != nil {
				return err
			}
		}

		// block until the user sends a CTRL+C
		select {}
	},
}

func startProxy(kashtiPort int) (*portforwarder.Tunnel, error) {
	config, err := getKubeConfig()
	if err != nil {
		return nil, err
	}

	c, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	kashtiSelector := labels.Set{"app": "kashti"}.AsSelector()
	tunnel, err := portforwarder.New(c, config, globalNamespace, kashtiSelector, remotePort, port)
	if err != nil {
		return nil, fmt.Errorf("cannot start port forward for kashti: %v", err)
	}

	return tunnel, nil
}
