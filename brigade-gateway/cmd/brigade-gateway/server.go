package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"gopkg.in/gin-gonic/gin.v1"
	"k8s.io/api/core/v1"

	"github.com/Azure/brigade/pkg/storage/kube"
	"github.com/Azure/brigade/pkg/webhook"
)

var (
	kubeconfig string
	master     string
	namespace  string
)

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "absolute path to the kubeconfig file")
	flag.StringVar(&master, "master", "", "master url")
	flag.StringVar(&namespace, "namespace", defaultNamespace(), "kubernetes namespace")
}

func main() {
	flag.Parse()

	clientset, err := kube.GetClient(master, kubeconfig)
	if err != nil {
		log.Fatal(err)
	}

	if namespace == "" {
		namespace = v1.NamespaceDefault
	}

	store := kube.New(clientset, namespace)

	router := gin.New()
	router.Use(gin.Recovery())

	events := router.Group("/events")
	{
		events.Use(gin.Logger())

		events.POST("/github", webhook.NewGithubHook(store).Handle)
		events.POST("/dockerhub/:org/:project/:commit", webhook.NewDockerPushHook(store).Handle)
	}

	router.GET("/healthz", healthz)

	router.Run(":7744")
}

func defaultNamespace() string {
	if ns, ok := os.LookupEnv("BRIGADE_NAMESPACE"); ok {
		return ns
	}
	return v1.NamespaceDefault
}

func healthz(c *gin.Context) {
	c.String(http.StatusOK, http.StatusText(http.StatusOK))
}
