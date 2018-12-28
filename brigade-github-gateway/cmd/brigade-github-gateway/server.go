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
	kubeconfig     string
	master         string
	namespace      string
	gatewayPort    string
	allowedAuthors authors
)

// defaultAllowedAuthors is the default set of authors allowed to PR
// https://developer.github.com/v4/reference/enum/commentauthorassociation/
var defaultAllowedAuthors = []string{"COLLABORATOR", "OWNER", "MEMBER"}

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "absolute path to the kubeconfig file")
	flag.StringVar(&master, "master", "", "master url")
	flag.StringVar(&namespace, "namespace", defaultNamespace(), "kubernetes namespace")
	flag.StringVar(&gatewayPort, "gateway-port", defaultGatewayPort(), "TCP port to use for brigade-github-gateway")
	flag.Var(&allowedAuthors, "authors", "allowed author associations, separated by commas (COLLABORATOR, CONTRIBUTOR, FIRST_TIMER, FIRST_TIME_CONTRIBUTOR, MEMBER, OWNER, NONE)")
}

func main() {
	flag.Parse()

	if len(allowedAuthors) == 0 {
		if aa, ok := os.LookupEnv("BRIGADE_AUTHORS"); ok {
			(&allowedAuthors).Set(aa)
		} else {
			allowedAuthors = defaultAllowedAuthors
		}
	}

	if len(allowedAuthors) > 0 {
		log.Printf("Forked PRs will be built for roles %s", strings.Join(allowedAuthors, " | "))
	}

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
		events.POST("/github", webhook.NewGithubHook(store, allowedAuthors))
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
	return "7744"
}

func healthz(c *gin.Context) {
	c.String(http.StatusOK, http.StatusText(http.StatusOK))
}

type authors []string

func (a *authors) Set(value string) error {
	for _, aa := range strings.Split(value, ",") {
		*a = append(*a, strings.ToUpper(aa))
	}
	return nil
}

func (a *authors) String() string {
	return strings.Join(*a, ",")
}
