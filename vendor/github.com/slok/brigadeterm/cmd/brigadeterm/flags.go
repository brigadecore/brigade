package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/alecthomas/kingpin.v2"
	"k8s.io/client-go/util/homedir"
)

const (
	brigadeNamespaceEnv = "BRIGADE_NAMESPACE"
)

type cmdFlags struct {
	app              *kingpin.Application
	kubeConfig       string
	kubeContext      string
	brigadeNamespace string
	showVersion      bool
	fake             bool
	reloadInterval   time.Duration
}

func newCmdFlags() (*cmdFlags, error) {
	fls := &cmdFlags{
		app: kingpin.New(filepath.Base(os.Args[0]), ""),
	}
	err := fls.init()

	return fls, err
}
func (c *cmdFlags) init() error {
	var kubehome string

	if kubehome = os.Getenv("KUBECONFIG"); kubehome == "" {
		kubehome = filepath.Join(homedir.HomeDir(), ".kube", "config")
	}

	// register flags
	c.app.Flag("kubeconfig", "Kubernetes configuration path").Default(kubehome).StringVar(&c.kubeConfig)
	c.app.Flag("namespace", "Kubernetes namespace where brigade is running (overrides $BRIGADE_NAMESPACE env var)").Short('n').Default("default").Envar(brigadeNamespaceEnv).StringVar(&c.brigadeNamespace)
	c.app.Flag("reload-interval", "The interval the UI will autoreload").Short('r').Default("3s").DurationVar(&c.reloadInterval)
	c.app.Flag("context", "Kubernetes context to use. Default to current context configured in kubeconfig").Short('c').Default("").StringVar(&c.kubeContext)
	c.app.Flag("version", "Show app version").Short('v').BoolVar(&c.showVersion)
	c.app.Flag("fake", "Run in fake mode").BoolVar(&c.fake)

	// Parse flags
	if _, err := c.app.Parse(os.Args[1:]); err != nil {
		return err
	}

	if err := c.validate(); err != nil {
		return err
	}

	return nil
}

func (c *cmdFlags) validate() error {
	if c.brigadeNamespace == "" {
		return fmt.Errorf("namespace is required")
	}
	return nil
}
