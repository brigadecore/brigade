package main

import (
	"flag"
	"log"
	"os"

	"gopkg.in/gin-gonic/gin.v1"
	"k8s.io/client-go/pkg/api/v1"

	"github.com/deis/brigade/pkg/api"
	"github.com/deis/brigade/pkg/storage/kube"
)

var (
	kubeconfig string
	master     string
	namespace  string
)

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "absolute path to the kubeconfig file")
	flag.StringVar(&master, "master", "", "master url")
	flag.StringVar(&namespace, "namespace", os.Getenv("BRIGADE_NAMESPACE"), "kubernetes namespace")
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

	storage := kube.New(clientset, namespace)

	router := gin.New()
	router.Use(gin.Recovery())

	// get an individual project
	rest := router.Group("/v1")
	rest.Use(gin.Logger(), cors)
	server := api.New(storage)

	p := server.Project()
	rest.GET("/projects", p.List)
	rest.GET("/project/:id", p.Get)
	rest.GET("/project/:id/builds", p.Builds)

	b := server.Build()
	rest.GET("/build/:id", b.Get)
	rest.GET("/build/:id/jobs", b.Jobs)

	j := server.Job()
	rest.GET("/job/:id", j.Get)
	rest.GET("/job/:id/logs", j.Logs)

	router.GET("/healthz", api.Healthz)
	log.Fatal(router.Run(":7745"))
}

func cors(c *gin.Context) {
	c.Header("access-control-allow-origin", "*")
}
