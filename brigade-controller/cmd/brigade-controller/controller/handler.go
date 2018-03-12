package controller

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"

	"k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Azure/brigade/pkg/storage/kube"
)

var (
	// ErrNoBuildID indicates that a secret does not have a build ID attached.
	ErrNoBuildID = errors.New("no build ID on secret")

	containerImageRegex = regexp.MustCompile("(.*):([^:]+)$")
)

func (c *Controller) syncSecret(secret *v1.Secret) error {
	// If a secret does not have a build ID then it cannot be tracked through
	// the system. A build ID should be a ULID.
	if bid, ok := secret.Labels["build"]; !ok || len(bid) == 0 {
		// Alternately, we could add a build ID and then re-save the secret.
		log.Printf("syncSecret: secret %s/%s has no build ID. Discarding.", secret.Namespace, secret.Name)
		return ErrNoBuildID
	}
	data := secret.Data

	log.Printf("EventHandler: type=%s provider=%s commit=%s", data["event_type"], data["event_provider"], data["commit_id"])

	podClient := c.clientset.CoreV1().Pods(secret.Namespace)

	if _, err := podClient.Get(secret.Name, metav1.GetOptions{}); err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}

		pid := secret.Labels["project"]
		if pid == "" {
			return errors.New("project ID not found")
		}

		secretClient := c.clientset.CoreV1().Secrets(secret.Namespace)
		project, err := secretClient.Get(pid, metav1.GetOptions{})
		if err != nil {
			return err
		}

		pod, err := c.newWorkerPod(secret, project)
		if err != nil {
			return err
		}
		if _, err := podClient.Create(&pod); err != nil {
			return err
		}
		log.Printf("Started %s for %q [%s] at %d", pod.Name, data["event_type"], data["commit_id"], pod.CreationTimestamp.Unix())
	}

	return nil
}

const (
	volumeName        = "brigade-build"
	volumeMountPath   = "/etc/brigade"
	sidecarVolumeName = "vcs-sidecar"
	sidecarVolumePath = "/vcs"
)

func (c *Controller) newWorkerPod(build, project *v1.Secret) (v1.Pod, error) {
	env := c.workerEnv(project, build)

	cmd := []string{"yarn", "-s", "start"}
	if cmdString, ok := project.Data["workerCommand"]; ok {
		cmd = strings.Split(string(cmdString), " ")
	}

	image, pullPolicy := c.workerImageConfig(project)

	podSpec := v1.PodSpec{
		ServiceAccountName: c.Config.WorkerServiceAccount,
		NodeSelector: map[string]string{
			"beta.kubernetes.io/os": "linux",
		},
		Containers: []v1.Container{{
			Name:            "brigade-runner",
			Image:           image,
			ImagePullPolicy: v1.PullPolicy(pullPolicy),
			Command:         cmd,
			VolumeMounts: []v1.VolumeMount{
				{
					Name:      volumeName,
					MountPath: volumeMountPath,
					ReadOnly:  true,
				},
				{
					Name:      sidecarVolumeName,
					MountPath: sidecarVolumePath,
					ReadOnly:  true,
				},
			},
			Env: env,
		}},
		Volumes: []v1.Volume{
			{
				Name: volumeName,
				VolumeSource: v1.VolumeSource{
					Secret: &v1.SecretVolumeSource{SecretName: build.Name},
				},
			},
			{
				Name: sidecarVolumeName,
				VolumeSource: v1.VolumeSource{
					EmptyDir: &v1.EmptyDirVolumeSource{},
				},
			},
		},
		RestartPolicy: v1.RestartPolicyNever,
	}

	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:   build.Name,
			Labels: build.Labels,
		},
		Spec: podSpec,
	}

	sv := kube.SecretValues(build.Data)

	// Skip adding the sidecar pod if the script is provided already.
	if len(sv.Bytes("script")) > 0 {
		return pod, nil
	}

	// Skip adding the sidecar pod if no sidecar pod image is supplied.
	if image := project.Data["vcsSidecar"]; len(image) > 0 {
		pod.Spec.InitContainers = []v1.Container{{
			Name:            "vcs-sidecar",
			Image:           string(image),
			ImagePullPolicy: v1.PullPolicy(pullPolicy),
			VolumeMounts: []v1.VolumeMount{{
				Name:      sidecarVolumeName,
				MountPath: sidecarVolumePath,
			}},
			Env: env,
		}}
	}
	return pod, nil
}

func (c *Controller) workerImageConfig(project *v1.Secret) (string, string) {
	// There isn't a correct way of making a proper distinction between registry,
	// registry+name or name, examples:
	//	* azure/brigade-worker:1234
	//	* myregisitry.com/azure/brigade-worker:1234
	// 	* myregistry/brigade-worker:1234
	// In order to tackle this, registry+name will be the name of the image.

	var name, tag string
	matches := containerImageRegex.FindStringSubmatch(c.WorkerImage)
	if len(matches) == 3 {
		name = matches[1]
		tag = matches[2]
	} else { // If no tag then name to default and tag to latest.
		name = c.WorkerImage
		tag = "latest"
	}

	sv := kube.SecretValues(project.Data)
	if n := sv.String("worker.name"); len(n) > 0 {
		name = n
	}
	if r := sv.String("worker.registry"); len(r) > 0 {
		// registry + name will work as name.
		name = fmt.Sprintf("%s/%s", r, name)
	}
	if t := sv.String("worker.tag"); len(t) > 0 {
		tag = t
	}
	image := fmt.Sprintf("%s:%s", name, tag)

	pullPolicy := c.WorkerPullPolicy
	if p := sv.String("worker.pullPolicy"); len(p) > 0 {
		pullPolicy = p
	}
	return image, pullPolicy
}

func (c *Controller) workerEnv(project, build *v1.Secret) []v1.EnvVar {
	sv := kube.SecretValues(build.Data)
	env := []v1.EnvVar{
		{Name: "CI", Value: "true"},
		{Name: "BRIGADE_BUILD_ID", Value: build.Labels["build"]},
		{Name: "BRIGADE_BUILD_NAME", Value: sv.String("build_name")},
		{Name: "BRIGADE_COMMIT_ID", Value: sv.String("commit_id")},
		{Name: "BRIGADE_COMMIT_REF", Value: sv.String("commit_ref")},
		{Name: "BRIGADE_EVENT_PROVIDER", Value: sv.String("event_provider")},
		{Name: "BRIGADE_EVENT_TYPE", Value: sv.String("event_type")},
		{Name: "BRIGADE_PROJECT_ID", Value: sv.String("project_id")},
		{Name: "BRIGADE_REMOTE_URL", Value: string(project.Data["cloneURL"])},
		{Name: "BRIGADE_WORKSPACE", Value: sidecarVolumePath},
		{
			Name: "BRIGADE_PROJECT_NAMESPACE",
			ValueFrom: &v1.EnvVarSource{
				FieldRef: &v1.ObjectFieldSelector{FieldPath: "metadata.namespace"},
			},
		},
		{
			Name:      "BRIGADE_REPO_KEY",
			ValueFrom: secretRef("sshKey", project),
		},
		{
			Name:      "BRIGADE_REPO_AUTH_TOKEN",
			ValueFrom: secretRef("github.token", project),
		},
		{Name: "BRIGADE_SERVICE_ACCOUNT", Value: c.Config.WorkerServiceAccount},
	}
	return env
}

// secretRef generate a SeccretKeyRef env var entry if `key` is present in `secret`.
// If the key does not exist a name/value pair is returned with an empty value
func secretRef(key string, secret *v1.Secret) *v1.EnvVarSource {
	trueVal := true
	return &v1.EnvVarSource{
		SecretKeyRef: &v1.SecretKeySelector{
			Key: key,
			LocalObjectReference: v1.LocalObjectReference{
				Name: secret.Name,
			},
			Optional: &trueVal,
		},
	}
}
