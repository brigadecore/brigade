package main

import (
	"flag"
	"log"

	"github.com/deis/acid/pkg/api"
	"github.com/deis/acid/pkg/config"
	"github.com/deis/acid/pkg/kube"
	"github.com/deis/acid/pkg/storage"

	"gopkg.in/gin-gonic/gin.v1"
)

func main() {
	kubeMaster := flag.String("master", "", "master url")
	kubeConfigLocation := flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	flag.Parse()
	clientset, err := kube.GetClient(*kubeMaster, *kubeConfigLocation)
	if err != nil {
		log.Fatalf("error creating kubernetes client (%s)", err)
		return
	}
	storage := storage.New(clientset)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())
	router.Use(config.Middleware())

	// get an individual project
	router.GET("/project/:id", api.Project(storage))

	router.GET("/healthz", api.Healthz)
	log.Fatal(router.Run(":7745"))
}
