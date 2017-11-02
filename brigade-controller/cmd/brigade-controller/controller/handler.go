package controller

import (
	"errors"
	"log"
	"strings"

	"k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Controller) syncSecret(secret *v1.Secret) error {
	data := secret.Data

	log.Printf("EventHandler: type=%s provider=%s commit=%s", data["event_type"], data["event_provider"], data["commit"])

	podClient := c.clientset.CoreV1().Pods(secret.Namespace)

	if _, err := podClient.Get(secret.Name, metav1.GetOptions{}); err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}

		pid := string(secret.Labels["project"])
		if pid == "" {
			return errors.New("project ID not found")
		}

		secretClient := c.clientset.CoreV1().Secrets(secret.Namespace)
		project, err := secretClient.Get(pid, metav1.GetOptions{})
		if err != nil {
			return err
		}

		pod, err := newWorkerPod(secret, project)
		if err != nil {
			return err
		}
		if _, err := podClient.Create(&pod); err != nil {
			return err
		}
		log.Printf("Started %s for %q [%s] at %d", pod.Name, data["event_type"], data["commit"], pod.CreationTimestamp.Unix())
	}

	return nil
}

const (
	brigadeWorkerImage      = "deis/brigade-worker:latest"
	brigadeWorkerPullPolicy = v1.PullIfNotPresent
	volumeName              = "brigade-build"
	volumeMountPath         = "/etc/brigade"
	sidecarVolumeName       = "vcs-sidecar"
	sidecarVolumePath       = "/vcs"
	vcsSidecarKey           = "vcsSidecar"
)

func newWorkerPod(secret, project *v1.Secret) (v1.Pod, error) {
	envvar := func(key string) v1.EnvVar {
		name := "BRIGADE_" + strings.ToUpper(key)
		return secretRef(name, key, secret)
	}

	podSpec := v1.PodSpec{
		Containers: []v1.Container{{
			Name:            "brigade-runner",
			Image:           brigadeWorkerImage,
			ImagePullPolicy: brigadeWorkerPullPolicy,
			Command:         []string{"yarn", "start"},
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
			Env: []v1.EnvVar{
				envvar("project_id"),
				envvar("event_type"),
				envvar("event_provider"),
				envvar("build_id"),
				envvar("commit"),
				{
					Name: "BRIGADE_PROJECT_NAMESPACE",
					ValueFrom: &v1.EnvVarSource{
						FieldRef: &v1.ObjectFieldSelector{FieldPath: "metadata.namespace"},
					},
				},
			},
		}},
		Volumes: []v1.Volume{
			{
				Name: volumeName,
				VolumeSource: v1.VolumeSource{
					Secret: &v1.SecretVolumeSource{SecretName: secret.Name},
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
			Name:   secret.Name,
			Labels: secret.Labels,
		},
		Spec: podSpec,
	}

	// Skip adding the sidecar pod if it's not necessary.
	if s, ok := secret.Data["script"]; ok && len(s) > 0 {
		return pod, nil
	}

	if image, ok := project.Data[vcsSidecarKey]; ok && len(image) > 0 {
		pod.Spec.InitContainers = []v1.Container{{
			Name:            "brigade-vcs-sidecar",
			Image:           string(image),
			ImagePullPolicy: brigadeWorkerPullPolicy,
			VolumeMounts: []v1.VolumeMount{{
				Name:      sidecarVolumeName,
				MountPath: sidecarVolumePath,
			}},
			Env: []v1.EnvVar{
				{
					Name:  "VCS_LOCAL_PATH",
					Value: sidecarVolumePath,
				},
				secretRef("VCS_REPO", "cloneURL", project),
				secretRef("VCS_REVISION", "commit", secret),
				secretRef("VCS_AUTH_TOKEN", "github.token", project),
				secretRef("BRIGADE_REPO_KEY", "sshKey", project),
			},
		}}
	}
	return pod, nil
}

// secretRef generate a SeccretKeyRef env var entry if `kye` is present in `secret`.
// If the key does not exist a name/value pair is returned with an empty value
func secretRef(name, key string, secret *v1.Secret) v1.EnvVar {
	if _, ok := secret.Data[key]; !ok {
		return v1.EnvVar{
			Name:  name,
			Value: "",
		}
	}
	return v1.EnvVar{
		Name: name,
		ValueFrom: &v1.EnvVarSource{
			SecretKeyRef: &v1.SecretKeySelector{
				Key: key,
				LocalObjectReference: v1.LocalObjectReference{
					Name: secret.Name,
				},
			},
		},
	}
}
