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

	handler := webhook.NewDockerPushHook(store)

	events := router.Group("/events")
	{
		events.Use(gin.Logger())

		// We need to handle the full project name (brigade-00000), the org/project
		// format of the name (for backward compatibility), and variants where the
		// commitish has to be supplied as a param.

		// Of the form /webhook/brigade-123456789?commit=master
		// Here, :org is actually a full project name, but due to Gin's naming rules
		// we have to keep it named :org.
		// This is the recommended form.
		events.POST("/webhook/:org", handler)

		// Of the form /webhook/deis/empty-testbed?commit=master
		events.POST("/webhook/:org/:repo", handler)
		// Of the form /webhook/deis/empty-testbed/master
		events.POST("/webhook/:org/:repo/:commit", handler)
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
