package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"gopkg.in/gin-gonic/gin.v1"
	"k8s.io/client-go/pkg/api/v1"

	"github.com/deis/acid/pkg/kube"
	"github.com/deis/acid/pkg/storage"
	"github.com/deis/acid/pkg/webhook"
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
		log.Fatal(err)
	}

	if namespace == "" {
		namespace = v1.NamespaceDefault
	}

	store := storage.New(clientset, namespace)

	router := gin.New()
	router.Use(gin.Recovery())

	events := router.Group("/events")
	{
		events.Use(gin.Logger())

		events.POST("/github", webhook.NewGithubHook(store).Handle)
		events.POST("/exec/:org/:project/:commit", webhook.NewExecHook(store).Handle)
		events.POST("/dockerhub/:org/:project/:commit", webhook.NewDockerPushHook(store).Handle)
	}

	// Lame UI
	ui := router.Group("/log/:org/:project")
	{
		ui.Use(gin.Logger())

		ui.GET("/", logHandler(clientset, store))
		ui.GET("/status.svg", badgeHandler(store))
		ui.GET("/id/:commit", logHandler(clientset, store))
	}

	router.GET("/healthz", healthz)

	router.Run(":7744")
}

func healthz(c *gin.Context) {
	c.String(http.StatusOK, http.StatusText(http.StatusOK))
}
