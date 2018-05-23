package commands

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"

	"github.com/Azure/brigade/pkg/portforwarder"
)

const (
	proxyUsage    = `Creates a tunnel to the Kashti component.`
	remoteAPIPort = 7745
	remotePort    = 80
)

var (
	port    int
	apiPort int

	kashtiNamespace string
)

func init() {
	Root.AddCommand(proxy)

	flags := proxy.PersistentFlags()
	flags.IntVar(&port, "port", 8081, "local port for the Kashti dashboard")
	flags.IntVar(&apiPort, "api-port", 7745, "local port for the Brigade API server")
	flags.StringVarP(&kashtiNamespace, "kashtiNamespace", "", "default", "namespace for Kashti")
}

var proxy = &cobra.Command{
	Use:   "proxy",
	Short: "Open a proxy to Kashti and the API server",
	Long:  proxyUsage,
	RunE: func(cmd *cobra.Command, args []string) error {
		return startProxy(port)
	},
}

func startProxy(kashtiPort int) error {

	config, err := getKubeConfig()
	if err != nil {
		return err
	}

	c, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	apiSelector := labels.Set{"role": "api"}.AsSelector()
	_, err = portforwarder.New(c, config, globalNamespace, apiSelector, remoteAPIPort, apiPort)
	if err != nil {
		return fmt.Errorf("cannot start port forward for brigade api: %v", err)
	}

	kashtiSelector := labels.Set{"app": "kashti"}.AsSelector()
	_, err = portforwarder.New(c, config, kashtiNamespace, kashtiSelector, remotePort, port)
	if err != nil {
		return fmt.Errorf("cannot start port forward for kashti: %v", err)
	}

	stop := make(chan os.Signal, 2)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-stop
		os.Exit(0)
	}()

	for {
		fmt.Printf("connect to kashti on http://localhost:%d\n", port)
		time.Sleep(10 * time.Second)
	}
}
