package controller

import (
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"regexp"
	"strings"

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apiresource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"strconv"

	"github.com/brigadecore/brigade/pkg/storage/kube"
)

var (
	// ErrNoBuildID indicates that a secret does not have a build ID attached.
	ErrNoBuildID = errors.New("no build ID on secret")

	containerImageRegex = regexp.MustCompile("(.*):([^:]+)$")
)

func (c *Controller) syncSecret(build *v1.Secret) error {
	// Ensure this secret has not yet been handled.
	if build.Labels["status"] == "accepted" {
		return nil
	}

	// If a secret does not have a build ID then it cannot be tracked through
	// the system. A build ID should be a ULID.
	if bid, ok := build.Labels["build"]; !ok || len(bid) == 0 {
		// Alternately, we could add a build ID and then re-save the secret.
		log.Printf("syncSecret: secret %s/%s has no build ID. Discarding.", build.Namespace, build.Name)
		return ErrNoBuildID
	}
	data := build.Data

	log.Printf("EventHandler: type=%s provider=%s commit=%s", data["event_type"], data["event_provider"], data["commit_id"])

	podClient := c.clientset.CoreV1().Pods(build.Namespace)

	if _, err := podClient.Get(build.Name, metav1.GetOptions{}); err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}

		pid := build.Labels["project"]
		if pid == "" {
			return errors.New("project ID not found")
		}

		secretClient := c.clientset.CoreV1().Secrets(build.Namespace)
		project, err := secretClient.Get(pid, metav1.GetOptions{})
		if err != nil {
			return err
		}

		pod := NewWorkerPod(build, project, c.Config)
		if _, err := podClient.Create(&pod); err != nil {
			return err
		}
		log.Printf("Started %s for %q [%s] at %d", pod.Name, data["event_type"], data["commit_id"], pod.CreationTimestamp.Unix())
	}

	return c.updateBuildStatus(build)
}

func (c *Controller) updateBuildStatus(build *v1.Secret) error {
	buildCopy := build.DeepCopy()
	buildCopy.Labels["status"] = "accepted"
	_, err := c.clientset.CoreV1().Secrets(build.Namespace).Update(buildCopy)
	return err
}

// NewWorkerPod returns pod context to create a worker pod
func NewWorkerPod(build, project *v1.Secret, config *Config) v1.Pod {
	env := workerEnv(project, build, config)

	cmd := []string{"yarn", "-s", "start"}
	if config.WorkerCommand != "" {
		cmd = strings.Split(config.WorkerCommand, " ")
	}
	if cmdBytes, ok := project.Data["workerCommand"]; ok && len(cmdBytes) > 0 {
		cmd = strings.Split(string(cmdBytes), " ")
	}

	image, pullPolicy := workerImageConfig(project, config)

	volumeMounts := []v1.VolumeMount{}
	buildVolumeMount := v1.VolumeMount{
		Name:      "brigade-build",
		MountPath: "/etc/brigade",
		ReadOnly:  true,
	}
	projectVolumeMount := v1.VolumeMount{
		Name:      "brigade-project",
		MountPath: "/etc/brigade-project",
		ReadOnly:  true,
	}
	sidecarVolumeMount := v1.VolumeMount{
		Name:      "vcs-sidecar",
		MountPath: "/vcs",
	}
	volumeMounts = append(volumeMounts, buildVolumeMount, projectVolumeMount)

	volumes := []v1.Volume{}
	buildVolume := v1.Volume{
		Name: buildVolumeMount.Name,
		VolumeSource: v1.VolumeSource{
			Secret: &v1.SecretVolumeSource{SecretName: build.Name},
		},
	}
	projectVolume := v1.Volume{
		Name: projectVolumeMount.Name,
		VolumeSource: v1.VolumeSource{
			Secret: &v1.SecretVolumeSource{SecretName: project.Name},
		},
	}
	sidecarVolume := v1.Volume{
		Name: sidecarVolumeMount.Name,
		VolumeSource: v1.VolumeSource{
			EmptyDir: &v1.EmptyDirVolumeSource{},
		},
	}
	volumes = append(volumes, buildVolume, projectVolume)

	initContainers := []v1.Container{}
	// Only add the sidecar resources if sidecar pod image is supplied.
	if image := project.Data["vcsSidecar"]; len(image) > 0 {
		volumeMounts = append(volumeMounts, sidecarVolumeMount)
		volumes = append(volumes, sidecarVolume)
		initContainers = append(initContainers,
			v1.Container{
				Name:            "vcs-sidecar",
				Image:           string(image),
				ImagePullPolicy: v1.PullPolicy(pullPolicy),
				VolumeMounts:    []v1.VolumeMount{sidecarVolumeMount},
				Env:             env,
				Resources:       vcsSidecarResources(project),
			})
	}

	spec := v1.PodSpec{
		ServiceAccountName: config.WorkerServiceAccount,
		NodeSelector: map[string]string{
			"beta.kubernetes.io/os": "linux",
		},
		Containers: []v1.Container{{
			Name:            "brigade-runner",
			Image:           image,
			ImagePullPolicy: v1.PullPolicy(pullPolicy),
			Command:         cmd,
			VolumeMounts:    volumeMounts,
			Env:             env,
			Resources:       workerResources(config),
		}},
		InitContainers: initContainers,
		Volumes:        volumes,
		RestartPolicy:  v1.RestartPolicyNever,
	}

	if scriptName := project.Data["defaultScriptName"]; len(scriptName) > 0 {
		attachConfigMap(&spec, string(scriptName), "/etc/brigade-default-script")
	}

	if ips := project.Data["imagePullSecrets"]; len(ips) > 0 {
		pullSecs := strings.Split(string(ips), ",")
		refs := []v1.LocalObjectReference{}
		for _, pullSec := range pullSecs {
			ref := v1.LocalObjectReference{Name: strings.TrimSpace(pullSec)}
			refs = append(refs, ref)
		}
		spec.ImagePullSecrets = refs
	}

	return v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:   build.Name,
			Labels: build.Labels,
		},
		Spec: spec,
	}
}

func workerImageConfig(project *v1.Secret, config *Config) (string, string) {
	// There isn't a correct way of making a proper distinction between registry,
	// registry+name or name, examples:
	//	* brigadecore/brigade-worker:1234
	//	* myregisitry.com/brigadecore/brigade-worker:1234
	// 	* myregistry/brigade-worker:1234
	// In order to tackle this, registry+name will be the name of the image.

	var name, tag string
	matches := containerImageRegex.FindStringSubmatch(config.WorkerImage)
	if len(matches) == 3 {
		name = matches[1]
		tag = matches[2]
	} else { // If no tag then name to default and tag to latest.
		name = config.WorkerImage
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

	pullPolicy := config.WorkerPullPolicy
	if p := sv.String("worker.pullPolicy"); len(p) > 0 {
		pullPolicy = p
	}
	return image, pullPolicy
}

func workerEnv(project, build *v1.Secret, config *Config) []v1.EnvVar {
	allowSecretKeyRef := false
	// older projects won't have allowSecretKeyRef set so just check for it
	if string(project.Data["kubernetes.allowSecretKeyRef"]) != "" {
		var err error
		allowSecretKeyRef, err = strconv.ParseBool(string(project.Data["kubernetes.allowSecretKeyRef"]))
		if err != nil {
			// if we errored parsing the bool something is wrong so just log it and ignore what the project set
			log.Printf("error parsing allowSecretKeyRef in project %s: %s", project.Annotations["projectName"], err)
		}
	}

	psv := kube.SecretValues(project.Data)
	bsv := kube.SecretValues(build.Data)

	serviceAccount := config.ProjectServiceAccount
	if string(project.Data["serviceAccount"]) != "" {
		serviceAccount = string(project.Data["serviceAccount"])
	}

	// Try to get cloneURL from the build first. This allows gateways to override
	// the project-level cloneURL if the commit that should be built, for
	// instance, exists only within a fork. If this isn't set at the build-level,
	// fall back to the project-level default.
	cloneURL := bsv.String("clone_url")
	if cloneURL == "" {
		cloneURL = string(project.Data["cloneURL"])
	}

	envs := []v1.EnvVar{
		{Name: "CI", Value: "true"},
		{Name: "BRIGADE_BUILD_ID", Value: build.Labels["build"]},
		{Name: "BRIGADE_BUILD_NAME", Value: bsv.String("build_name")},
		{Name: "BRIGADE_COMMIT_ID", Value: bsv.String("commit_id")},
		{Name: "BRIGADE_COMMIT_REF", Value: bsv.String("commit_ref")},
		{Name: "BRIGADE_EVENT_PROVIDER", Value: bsv.String("event_provider")},
		{Name: "BRIGADE_EVENT_TYPE", Value: bsv.String("event_type")},
		{Name: "BRIGADE_PROJECT_ID", Value: bsv.String("project_id")},
		{Name: "BRIGADE_LOG_LEVEL", Value: bsv.String("log_level")},
		{Name: "BRIGADE_REMOTE_URL", Value: cloneURL},
		{Name: "BRIGADE_WORKSPACE", Value: "/vcs"},
		{Name: "BRIGADE_PROJECT_NAMESPACE", Value: build.Namespace},
		{Name: "BRIGADE_SERVICE_ACCOUNT", Value: serviceAccount},
		{Name: "BRIGADE_SECRET_KEY_REF", Value: strconv.FormatBool(allowSecretKeyRef)},
		{
			Name:      "BRIGADE_REPO_KEY",
			ValueFrom: secretRef("sshKey", project),
		}, {
			Name:      "BRIGADE_REPO_AUTH_TOKEN",
			ValueFrom: secretRef("github.token", project),
		},
		{Name: "BRIGADE_DEFAULT_BUILD_STORAGE_CLASS", Value: config.DefaultBuildStorageClass},
		{Name: "BRIGADE_DEFAULT_CACHE_STORAGE_CLASS", Value: config.DefaultCacheStorageClass},
	}

	if config.ProjectServiceAccountRegex != "" {
		envs = append(envs, v1.EnvVar{Name: "BRIGADE_SERVICE_ACCOUNT_REGEX", Value: config.ProjectServiceAccountRegex})
	}

	brigadejsPath := psv.String("brigadejsPath")
	if brigadejsPath != "" {
		if filepath.IsAbs(brigadejsPath) {
			log.Printf("Warning: 'brigadejsPath' is set on Project Secret but will be ignored because provided path '%s' is an absolute path", brigadejsPath)
		} else {
			envs = append(envs, v1.EnvVar{Name: "BRIGADE_SCRIPT", Value: filepath.Join("/vcs", brigadejsPath)})
		}
	}

	return envs
}

// workerResources generates the resources for the worker, given in the cofiguration
// If the value is not given, or it's wrong, empty resources gill be returned
func workerResources(config *Config) v1.ResourceRequirements {
	resources := v1.ResourceRequirements{
		Limits:   v1.ResourceList{},
		Requests: v1.ResourceList{},
	}

	if v, err := apiresource.ParseQuantity(config.WorkerLimitsCPU); err == nil {
		resources.Limits[v1.ResourceCPU] = v
	}
	if v, err := apiresource.ParseQuantity(config.WorkerLimitsMemory); err == nil {
		resources.Limits[v1.ResourceMemory] = v
	}
	if v, err := apiresource.ParseQuantity(config.WorkerRequestsCPU); err == nil {
		resources.Requests[v1.ResourceCPU] = v
	}
	if v, err := apiresource.ParseQuantity(config.WorkerRequestsMemory); err == nil {
		resources.Requests[v1.ResourceMemory] = v
	}

	return resources
}

// vcsSidecarResources generates the resources for the init-container in the worker
// If the value is not given, or it's wrong, empty resources gill be returned
func vcsSidecarResources(project *v1.Secret) v1.ResourceRequirements {
	resources := v1.ResourceRequirements{
		Limits:   v1.ResourceList{},
		Requests: v1.ResourceList{},
	}

	if givenValue, ok := project.Data["vcsSidecarResources.limits.cpu"]; ok {
		if v, err := apiresource.ParseQuantity(string(givenValue)); err == nil {
			resources.Limits[v1.ResourceCPU] = v
		}
	}
	if givenValue, ok := project.Data["vcsSidecarResources.limits.memory"]; ok {
		if v, err := apiresource.ParseQuantity(string(givenValue)); err == nil {
			resources.Limits[v1.ResourceMemory] = v
		}
	}
	if givenValue, ok := project.Data["vcsSidecarResources.requests.cpu"]; ok {
		if v, err := apiresource.ParseQuantity(string(givenValue)); err == nil {
			resources.Requests[v1.ResourceCPU] = v
		}
	}
	if givenValue, ok := project.Data["vcsSidecarResources.requests.memory"]; ok {
		if v, err := apiresource.ParseQuantity(string(givenValue)); err == nil {
			resources.Requests[v1.ResourceMemory] = v
		}
	}

	return resources
}

// secretRef generate a SecretKeyRef env var entry if `key` is present in `secret`.
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

func attachConfigMap(spec *v1.PodSpec, name, path string) {
	spec.Volumes = append(spec.Volumes, v1.Volume{
		Name: name,
		VolumeSource: v1.VolumeSource{
			ConfigMap: &v1.ConfigMapVolumeSource{
				LocalObjectReference: v1.LocalObjectReference{
					Name: name,
				},
			},
		},
	})

	for i := range spec.Containers {
		spec.Containers[i].VolumeMounts = append(
			spec.Containers[i].VolumeMounts,
			v1.VolumeMount{
				Name:      name,
				MountPath: path,
			})
	}
}
