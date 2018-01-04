package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	gin "gopkg.in/gin-gonic/gin.v1"
	"k8s.io/api/core/v1"

	"github.com/Azure/brigade/pkg/api"
	"github.com/Azure/brigade/pkg/storage/kube"
)

var (
	kubeconfig string
	master     string
	namespace  string
	apiPort    string
)

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "absolute path to the kubeconfig file")
	flag.StringVar(&master, "master", "", "master url")
	flag.StringVar(&namespace, "namespace", defaultNamespace(), "kubernetes namespace")
	flag.StringVar(&apiPort, "api-port", defaultAPIPort(), "TCP port to use for brigade-api")
}

func main() {
	flag.Parse()
	clientset, err := kube.GetClient(master, kubeconfig)
	if err != nil {
		log.Fatalf("error creating kubernetes client (%s)", err)
		return
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
	rest.GET("/projects-build", p.ListWithLatestBuild)

	b := server.Build()
	rest.GET("/build/:id", b.Get)
	rest.GET("/build/:id/jobs", b.Jobs)
	rest.GET("/build/:id/logs", b.Logs)

	j := server.Job()
	rest.GET("/job/:id", j.Get)
	rest.GET("/job/:id/logs", j.Logs)

	router.GET("/healthz", api.Healthz)

	formattedAPIPort := fmt.Sprintf(":%v", apiPort)
	log.Fatal(router.Run(formattedAPIPort))
}

func defaultNamespace() string {
	if ns, ok := os.LookupEnv("BRIGADE_NAMESPACE"); ok {
		return ns
	}
	return v1.NamespaceDefault
}

func defaultAPIPort() string {
	if port, ok := os.LookupEnv("BRIGADE_API_PORT"); ok {
		return port
	}
	return "7745"
}

func cors(c *gin.Context) {
	c.Header("access-control-allow-origin", "*")
}
