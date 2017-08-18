package controller

import (
	"log"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/api/v1"
)

func (c *Controller) syncSecret(secret *v1.Secret) error {
	data := secret.Data

	log.Printf("EventHandler: type=%s provider=%s commit=%s", data["event_type"], data["event_provider"], data["commit"])

	podClient := c.clientset.CoreV1().Pods(secret.Namespace)

	if _, err := podClient.Get(secret.Name, metav1.GetOptions{}); err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}

		pod := newWorkerPod(secret)
		if _, err := podClient.Create(&pod); err != nil {
			return err
		}
		log.Printf("Started %s for %q [%s] at %d", pod.Name, data["event_type"], data["commit"], pod.CreationTimestamp.Unix())
	}

	return nil
}

const (
	acidWorkerImage      = "acidic.azurecr.io/acid-worker:latest"
	acidWorkerPullPolicy = v1.PullIfNotPresent
	volumeName           = "acid-build"
	volumeMountPath      = "/etc/acid"
)

func newWorkerPod(secret *v1.Secret) v1.Pod {
	envvar := func(key string) v1.EnvVar {
		return v1.EnvVar{
			Name: "ACID_" + strings.ToUpper(key),
			ValueFrom: &v1.EnvVarSource{
				SecretKeyRef: &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: secret.Name,
					},
					Key: key,
				},
			},
		}
	}

	podSpec := v1.PodSpec{
		Containers: []v1.Container{{
			Name:            "acid-runner",
			Image:           acidWorkerImage,
			ImagePullPolicy: acidWorkerPullPolicy,
			Command:         []string{"yarn", "start"},
			VolumeMounts: []v1.VolumeMount{{
				Name:      volumeName,
				MountPath: volumeMountPath,
				ReadOnly:  true,
			}},
			Env: []v1.EnvVar{
				envvar("project_id"),
				envvar("event_type"),
				envvar("event_provider"),
				envvar("commit"),
				{
					Name: "ACID_PROJECT_NAMESPACE",
					ValueFrom: &v1.EnvVarSource{
						FieldRef: &v1.ObjectFieldSelector{FieldPath: "metadata.namespace"},
					},
				},
			},
		}},
		Volumes: []v1.Volume{{
			Name: volumeName,
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{SecretName: secret.Name},
			}},
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
	return pod
}
