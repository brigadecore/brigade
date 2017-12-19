package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	gin "gopkg.in/gin-gonic/gin.v1"
	"k8s.io/api/core/v1"

	"github.com/Azure/brigade/pkg/storage/kube"
	"github.com/Azure/brigade/pkg/webhook"
)

var (
	kubeconfig              string
	master                  string
	namespace               string
	gatewayPort             string
	buildForkedPullRequests bool
)

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "absolute path to the kubeconfig file")
	flag.StringVar(&master, "master", "", "master url")
	flag.StringVar(&namespace, "namespace", defaultNamespace(), "kubernetes namespace")
	flag.StringVar(&gatewayPort, "gateway-port", defaultGatewayPort(), "TCP port to use for brigade-gateway")
	flag.BoolVar(&buildForkedPullRequests, "build-forked-pull-requests", defaultBuildForkedPRs(), "build forked pull requests")
}

func main() {
	flag.Parse()

	clientset, err := kube.GetClient(master, kubeconfig)
	if err != nil {
		log.Fatal(err)
	}

	store := kube.New(clientset, namespace)

	router := gin.New()
	router.Use(gin.Recovery())

	events := router.Group("/events")
	{
		events.Use(gin.Logger())
		events.POST("/github", webhook.NewGithubHook(store, buildForkedPullRequests).Handle)
	}

	router.GET("/healthz", healthz)

	formattedGatewayPort := fmt.Sprintf(":%v", gatewayPort)
	router.Run(formattedGatewayPort)
}

func defaultNamespace() string {
	if ns, ok := os.LookupEnv("BRIGADE_NAMESPACE"); ok {
		return ns
	}
	return v1.NamespaceDefault
}

func defaultGatewayPort() string {
	if port, ok := os.LookupEnv("BRIGADE_GATEWAY_PORT"); ok {
		return port
	}
	return "7745"
}

func defaultBuildForkedPRs() bool {
	if v, ok := os.LookupEnv("BRIGADE_BUILD_FORKED_PULL_REQUESTS"); ok {
		if v == "1" || strings.EqualFold(v, "true") {
			return true
		}
	}
	return false
}

func healthz(c *gin.Context) {
	c.String(http.StatusOK, http.StatusText(http.StatusOK))
}
