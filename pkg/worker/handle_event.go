package worker

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/api/v1"
)

// DefaultExecutor handles the event.
// By default, we configure to use the Kubernetes executor.
var DefaultExecutor Executor = &k8sExecutor{}

// Pod defaults for the acid worker.
var (
	DefaultAcidWorker               = "acidic.azurecr.io/acid-worker:latest"
	DefaultPullPolicy v1.PullPolicy = "IfNotPresent"
)

// HandleEvent creates a default sandbox and then executes the given Acid.js for the given event.
func HandleEvent(e *Event, proj *Project, acidjs []byte) error {

	pod := runnerPod(e, proj, acidjs)
	// Execute pod
	_, err := DefaultExecutor.Create(proj.Kubernetes.Namespace, &pod)

	// TODO: Probably we want to return the final pod spec (_ above) so that
	// something else can watch the pod. Alternatively, we can probably return
	// just the pod name or ID.
	log.Printf("Started %s for %s at %d %s", pod.Name, e.Type, pod.CreationTimestamp.Unix(), err)

	return err
}

// Executor takes a pod and handles inserting it into a runtime environment.
type Executor interface {
	Create(namespace string, pod *v1.Pod) (*v1.Pod, error)
}

// runnerPod creates a Pod definition for a runner.
func runnerPod(e *Event, proj *Project, acidjs []byte) v1.Pod {
	shortCommit := e.Commit
	if len(shortCommit) > 8 {
		shortCommit = shortCommit[0:8]
	}

	jobName := fmt.Sprintf("acid-worker-%d-%s", time.Now().Unix(), shortCommit)
	cleanProjName := strings.Replace(proj.Repo.Name, "/", "-", -1)

	// TODO: Are we only passing []byte in payloads now?
	payload, err := json.Marshal(e.Payload)
	if err != nil {
		panic(err)
	}

	encodedScript := base64.StdEncoding.EncodeToString(acidjs)

	return v1.Pod{
		ObjectMeta: meta.ObjectMeta{
			Name: jobName,
			Labels: map[string]string{
				"heritage":  "acid",
				"managedBy": "acid",
				"jobname":   jobName,
				"belongsto": cleanProjName,
				"commit":    e.Commit,
				// Need a different name for this.
				"role": "build",
			},
			Annotations: map[string]string{},
		},
		Spec: v1.PodSpec{
			RestartPolicy: "Never",
			Containers: []v1.Container{
				{
					Name:  "acid-worker",
					Image: DefaultAcidWorker,
					Command: []string{
						"npm", "start",
					},
					ImagePullPolicy: DefaultPullPolicy,
					Env: []v1.EnvVar{
						{
							Name:  "ACID_EVENT_TYPE",
							Value: e.Type,
						},
						{
							Name:  "ACID_EVENT_PROVIDER",
							Value: e.Provider,
						},
						{
							Name:  "ACID_COMMIT",
							Value: e.Commit,
						},
						{
							Name:  "ACID_PAYLOAD",
							Value: string(payload),
						},
						{
							Name:  "ACID_PROJECT_ID",
							Value: proj.ID,
						},
						{
							Name:  "ACID_PROJECT_NAMESPACE",
							Value: proj.Kubernetes.Namespace,
						},
						{
							Name:  "ACID_SCRIPT",
							Value: encodedScript,
						},
					},
				},
			},
			Volumes: []v1.Volume{},
		},
	}
}
