package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"k8s.io/api/core/v1"

	"github.com/uswitch/brigade/pkg/api"
	"github.com/uswitch/brigade/pkg/storage/kube"
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
	router.Use(cors.Default())

	// get an individual project
	rest := router.Group("/v1")
	server := api.New(storage)

	p := server.Project()
	rest.GET("/projects", gin.HandlerFunc(p.List))
	rest.GET("/project/:id", gin.HandlerFunc(p.Get))
	rest.GET("/project/:id/builds", gin.HandlerFunc(p.Builds))
	rest.GET("/projects-build", gin.HandlerFunc(p.ListWithLatestBuild))

	b := server.Build()
	rest.GET("/build/:id", gin.HandlerFunc(b.Get))
	rest.GET("/build/:id/jobs", gin.HandlerFunc(b.Jobs))
	rest.GET("/build/:id/logs", gin.HandlerFunc(b.Logs))

	j := server.Job()
	rest.GET("/job/:id", gin.HandlerFunc(j.Get))
	rest.GET("/job/:id/logs", gin.HandlerFunc(j.Logs))

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
