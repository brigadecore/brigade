package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"gopkg.in/gin-gonic/gin.v1"
	"k8s.io/api/core/v1"

	"github.com/Azure/brigade/pkg/storage"
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

	router := newRouter(store)
	router.Run(":8000")
}

func newRouter(store storage.Store) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())

	events := router.Group("/events")
	{
		events.Use(gin.Logger())

		events.POST("/webhook/:org/:project/:commit", webhook.NewDockerPushHook(store).Handle)
	}

	router.GET("/healthz", healthz)

	return router
}

func healthz(c *gin.Context) {
	c.String(http.StatusOK, http.StatusText(http.StatusOK))
}

func defaultNamespace() string {
	if ns, ok := os.LookupEnv("BRIGADE_NAMESPACE"); ok {
		return ns
	}
	return v1.NamespaceDefault
}
