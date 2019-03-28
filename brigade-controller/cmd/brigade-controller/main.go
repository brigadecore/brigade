package main

import (
	"flag"
	"log"
	"os"

	"github.com/brigadecore/brigade/brigade-controller/cmd/brigade-controller/controller"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const defaultWorkerServiceAccountName = "brigade-worker"
const defaultJobServiceAccountName = "brigade-worker"

func init() {
	log.SetFlags(log.Lshortfile)
}

func main() {
	var (
		kubeconfig string
		master     string
		ctrConfig  controller.Config
	)

	flag.StringVar(&kubeconfig, "kubeconfig", "", "absolute path to the kubeconfig file")
	flag.StringVar(&master, "master", "", "master url")
	flag.StringVar(&ctrConfig.Namespace, "namespace", defaultNamespace(), "kubernetes namespace")
	flag.StringVar(&ctrConfig.WorkerImage, "worker-image", defaultWorkerImage(), "kubernetes worker image")
	flag.StringVar(&ctrConfig.WorkerPullPolicy, "worker-pull-policy", defaultWorkerPullPolicy(), "kubernetes worker image pull policy")
	flag.StringVar(&ctrConfig.WorkerServiceAccount, "worker-service-account", defaultWorkerServiceAccount(), "kubernetes worker service account name")
	flag.StringVar(&ctrConfig.ProjectServiceAccount, "project-service-account", defaultProjectServiceAccount(), "default brigade project service account name")
	flag.StringVar(&ctrConfig.ProjectServiceAccountRegex, "project-service-account-regex", "", "regex to validate project service accounts, if not given will be set to match the default project service account")
	flag.StringVar(&ctrConfig.WorkerRequestsCPU, "worker-requests-cpu", "", "kubernetes worker cpu requests")
	flag.StringVar(&ctrConfig.WorkerRequestsMemory, "worker-requests-memory", "", "kubernetes worker memory requests")
	flag.StringVar(&ctrConfig.WorkerLimitsCPU, "worker-limits-cpu", "", "kubernetes worker cpu limits")
	flag.StringVar(&ctrConfig.WorkerLimitsMemory, "worker-limits-memory", "", "kubernetes worker memory limits")
	flag.Parse()

	if ctrConfig.ProjectServiceAccountRegex == "" {
		// No regex was given so only allow the default project service account
		ctrConfig.ProjectServiceAccountRegex = ctrConfig.ProjectServiceAccount
	}

	// creates the connection
	config, err := clientcmd.BuildConfigFromFlags(master, kubeconfig)
	if err != nil {
		log.Fatal(err)
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	controller := controller.NewController(clientset, &ctrConfig)
	log.Printf("Listening in namespace %q for new events", ctrConfig.Namespace)

	// Now let's start the controller
	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)

	// Wait forever
	select {}
}

func defaultWorkerImage() string {
	if image, ok := os.LookupEnv("BRIGADE_WORKER_IMAGE"); ok {
		return image
	}
	return "brigadecore/brigade-worker:latest"
}

func defaultWorkerPullPolicy() string {
	if pp, ok := os.LookupEnv("BRIGADE_WORKER_PULL_POLICY"); ok {
		return pp
	}
	return string(v1.PullIfNotPresent)
}

func defaultWorkerServiceAccount() string {
	if pp, ok := os.LookupEnv("BRIGADE_WORKER_SERVICE_ACCOUNT"); ok {
		return pp
	}
	return defaultWorkerServiceAccountName
}

func defaultProjectServiceAccount() string {
	if pp, ok := os.LookupEnv("BRIGADE_JOB_SERVICE_ACCOUNT"); ok {
		return pp
	}
	return defaultJobServiceAccountName
}

func defaultNamespace() string {
	if ns, ok := os.LookupEnv("BRIGADE_NAMESPACE"); ok {
		return ns
	}
	return v1.NamespaceDefault
}
