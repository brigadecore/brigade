package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	apps_v1 "k8s.io/api/apps/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	Root.AddCommand(check)
}

const (
	checkUsage = `Checks the status of your Brigade installation

Specifically, it reports Desired/Current/Running/Up-to-date/Available/Unavailable Pods for the Controller, API Server and Kashti deployments.
`
)

var check = &cobra.Command{
	Use:   "check",
	Short: "Checks the status of your Brigade installation",
	Long:  checkUsage,
	RunE: func(cmd *cobra.Command, args []string) error {
		return checkBrigadeSystem()
	},
}

func checkBrigadeSystem() error {
	/*
		If you run `kubectl get deploy --show-labels` on the namespace where Brigade is installed, you'll see
		NAME                                DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE    LABELS
		mybrigade-brigade-api               1         1         1            1           170m   app.kubernetes.io/instance=mybrigade,app.kubernetes.io/managed-by=Tiller,app.kubernetes.io/name=mybrigade-brigade-api,helm.sh/chart=brigade-1.0.0,role=api
		mybrigade-brigade-ctrl              1         1         1            1           170m   app.kubernetes.io/instance=mybrigade,app.kubernetes.io/managed-by=Tiller,app.kubernetes.io/name=mybrigade-brigade-ctrl,helm.sh/chart=brigade-1.0.0,role=controller
		mybrigade-brigade-generic-gateway   1         1         1            1           170m   app.kubernetes.io/instance=mybrigade,app.kubernetes.io/managed-by=Tiller,app.kubernetes.io/name=mybrigade-brigade,helm.sh/chart=brigade-1.0.0,role=gateway,type=generic
		mybrigade-kashti                    1         1         1            1           170m   app=kashti,chart=kashti-0.1.1,heritage=Tiller,release=mybrigade
	*/

	c, err := kubeClient()
	if err != nil {
		return err
	}

	// check deployments
	deployList, err := c.AppsV1().Deployments(globalNamespace).List(context.TODO(), meta_v1.ListOptions{})
	if err != nil {
		return err
	}

	apiDeploymentFound := false
	controllerDeploymentFound := false
	kashtiDeploymentFound := false

	for _, deployment := range deployList.Items {
		if lbl := deployment.Labels["app.kubernetes.io/name"]; lbl != "" {
			if strings.HasSuffix(lbl, "-brigade-api") && deployment.Labels["role"] == "api" {
				apiDeploymentFound = true
				reportDeployStatus(deployment, "Brigade API Server")
			} else if strings.HasSuffix(lbl, "-brigade-ctrl") && deployment.Labels["role"] == "controller" {
				controllerDeploymentFound = true
				reportDeployStatus(deployment, "Brigade Controller")
			}
		} else if lblApp := deployment.Labels["app"]; lblApp == "kashti" {
			kashtiDeploymentFound = true
			reportDeployStatus(deployment, "Kashti")
		}

		if apiDeploymentFound && controllerDeploymentFound && kashtiDeploymentFound {
			break // we're not interested in checking other Deployments
		}
	}

	if !apiDeploymentFound {
		fmt.Printf("Info: Brigade API Server Deployment not found")
	}
	if !controllerDeploymentFound {
		fmt.Println("Error: Brigade Controller Deployment not found")
	}
	if !kashtiDeploymentFound {
		fmt.Println("Info: Kashti deployment not found")
	}

	// check vacuum cronjob
	cjList, err := c.BatchV1beta1().CronJobs(globalNamespace).List(context.TODO(), meta_v1.ListOptions{})
	if err != nil {
		return err
	}
	vacuumCronJobFound := false
	for _, cronjob := range cjList.Items {
		if lbl := cronjob.Labels["app.kubernetes.io/name"]; lbl != "" {
			if strings.HasSuffix(lbl, "-brigade") && cronjob.Labels["role"] == "vacuum" {
				vacuumCronJobFound = true
				if *cronjob.Spec.Suspend {
					fmt.Println("Warning: Vacuum CronJob is suspended")
				} else if cronjob.Spec.Schedule == "" {
					fmt.Println("Warning: Vacuum has an empty schedule")
				} else {
					fmt.Println("Info: Vacuum is healthy (not suspended and its schedule is non-empty)")
				}
			}
		}

		if vacuumCronJobFound {
			break // we're not interested in checking other CronJobs
		}
	}

	return nil
}

func reportDeployStatus(deployment apps_v1.Deployment, name string) {
	fmt.Print(deployment, name)
}

func getDeployStatusString(deployment apps_v1.Deployment, name string) string {
	ds := deployment.Status

	var emoji string

	if ds.Replicas == ds.ReadyReplicas {
		emoji = "✔️"
	} else {
		emoji = "❌"
	}

	return fmt.Sprintf("%s %s replicas: Desired %d, Ready %d, Updated %d, Available %d, Unavailable %d \n", emoji, name, ds.Replicas, ds.ReadyReplicas, ds.UpdatedReplicas, ds.AvailableReplicas, ds.UnavailableReplicas)
}
