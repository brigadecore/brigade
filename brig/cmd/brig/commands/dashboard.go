package commands

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bacongobbler/browser"
	"github.com/spf13/cobra"
)

const dashboardUsage = `Opens the Kashti dashboard with the default browser.
`

func init() {
	Root.AddCommand(dashboard)

	flags := dashboard.PersistentFlags()
	flags.IntVar(&port, "port", 8081, "local port for the Kashti dashboard")
	flags.IntVar(&apiPort, "api-port", 7745, "local port for the Brigade API server")
	flags.StringVarP(&kashtiNamespace, "kashti-namespace", "", "default", "namespace for Kashti")
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

		log.Printf("Connected! When you are finished with this session, enter CTRL+C.", port)

		if err := browser.Open(fmt.Sprintf("http://localhost:%d", port)); err != nil {
			return err
		}

		// block until the user sends a CTRL+C
		for {
		}
	},
}
