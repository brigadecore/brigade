package main

import (
	"flag"
	"log"
	"os"

	"github.com/Azure/brigade/brigade-controller/cmd/brigade-controller/controller"

	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func init() {
	log.SetFlags(log.Lshortfile)
}

func main() {
	var (
		kubeconfig string
		master     string
		ctrConfig  controller.Config
	)

	flag.StringVar(&kubeconfig, "kubeconfig", "", "absolute path to the kubeconfig file")
	flag.StringVar(&master, "master", "", "master url")
	flag.StringVar(&ctrConfig.Namespace, "namespace", defaultNS(), "kubernetes namespace")
	flag.StringVar(&ctrConfig.WorkerImage, "worker-image", os.Getenv("BRIGADE_WORKER_IMAGE"), "kubernetes namespace")
	flag.StringVar(&ctrConfig.WorkerPullPolicy, "worker-pull-policy", os.Getenv("BRIGADE_WORKER_PULL_POLICY"), "kubernetes namespace")
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

	controller := controller.NewController(clientset, &ctrConfig)
	log.Printf("Listening in namespace %q for new events", ctrConfig.Namespace)

	// Now let's start the controller
	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)

	// Wait forever
	select {}
}

func defaultNS() string {
	ns := os.Getenv("BRIGADE_NAMESPACE")
	if ns == "" {
		return v1.NamespaceDefault
	}
	return ns
}
