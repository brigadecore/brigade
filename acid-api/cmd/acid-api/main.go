package main

import (
	"flag"
	"log"
	"os"

	"gopkg.in/gin-gonic/gin.v1"
	"k8s.io/client-go/pkg/api/v1"

	"github.com/deis/acid/pkg/api"
	"github.com/deis/acid/pkg/kube"
	"github.com/deis/acid/pkg/storage"
)

var (
	kubeconfig string
	master     string
	namespace  string
)

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "absolute path to the kubeconfig file")
	flag.StringVar(&master, "master", "", "master url")
	flag.StringVar(&namespace, "namespace", os.Getenv("ACID_NAMESPACE"), "kubernetes namespace")
}

func main() {
	flag.Parse()
	clientset, err := kube.GetClient(master, kubeconfig)
	if err != nil {
		log.Fatalf("error creating kubernetes client (%s)", err)
		return
	}

	if namespace == "" {
		namespace = v1.NamespaceDefault
	}

	storage := storage.New(clientset, namespace)

	router := gin.New()
	router.Use(gin.Recovery())

	// get an individual project
	rest := router.Group("/v1")
	{
		rest.Use(gin.Logger())
		rest.GET("/project/:id", api.Project(storage))
		rest.GET("/build/:id", api.Build(storage))
	}

	router.GET("/healthz", api.Healthz)
	log.Fatal(router.Run(":7745"))
}
