package main

import (
	"flag"
	"log"
	"os"

	"github.com/deis/acid/acid-controller/cmd/acid-controller/controller"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/tools/clientcmd"
)

func init() {
	log.SetFlags(log.Lshortfile)
}

func main() {
	var (
		kubeconfig string
		master     string
		namespace  string
	)

	flag.StringVar(&kubeconfig, "kubeconfig", "", "absolute path to the kubeconfig file")
	flag.StringVar(&master, "master", "", "master url")
	flag.StringVar(&namespace, "namespace", defaultNS(), "kubernetes namespace")
	flag.Parse()

	// creates the connection
	config, err := clientcmd.BuildConfigFromFlags(master, kubeconfig)
	if err != nil {
		log.Fatal(err)
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	controller := controller.NewController(clientset, namespace)
	log.Printf("Listening in namespace %q for new events", namespace)

	// Now let's start the controller
	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)

	// Wait forever
	select {}
}

func defaultNS() string {
	ns := os.Getenv("ACID_NAMESPACE")
	if ns == "" {
		return v1.NamespaceDefault
	}
	return ns
}
