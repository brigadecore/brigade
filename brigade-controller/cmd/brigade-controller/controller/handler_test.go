package controller

import (
	"strings"
	"testing"

	v1 "k8s.io/api/core/v1"
)

func TestNewWorkerPod_Defaults(t *testing.T) {
	build := &v1.Secret{}
	proj := &v1.Secret{
		Data: map[string][]byte{
			"vcsSidecar": []byte("my-vcs-sidecar"),
		},
	}
	config := &Config{
		Namespace: v1.NamespaceDefault,
	}

	pod := NewWorkerPod(build, proj, config)

	spec := pod.Spec
	if spec.NodeSelector["beta.kubernetes.io/os"] != "linux" {
		t.Error("expected linux node selector")
	}
	sidecarVolumeExists := false
	for _, volume := range spec.Volumes {
		if volume.Name == "vcs-sidecar" {
			sidecarVolumeExists = true
		}
	}
	if !sidecarVolumeExists {
		t.Error("expected vcs-sidecar volume to exist")
	}
	if len(spec.InitContainers) == 0 {
		t.Error("expected spec.InitContainers to be non-zero")
	}

	container := spec.Containers[0]
	if container.Name != "brigade-runner" {
		t.Error("expected brigade-runner container name")
	}

	sidecarVolumeMountExists := false
	for _, volumeMount := range container.VolumeMounts {
		if volumeMount.Name == "vcs-sidecar" {
			sidecarVolumeMountExists = true
		}
	}
	if !sidecarVolumeMountExists {
		t.Error("Expected vcs-sidecar volume mount to exist")
	}

	if cmd := strings.Join(container.Command, " "); cmd != "yarn -s start" {
		t.Errorf("Unexpected command: %s", cmd)
	}

	if len(container.Resources.Limits) != 0 {
		t.Errorf("Limits should be undefined")
	}

	if len(container.Resources.Requests) != 0 {
		t.Errorf("Requests should be undefined")
	}
}

func TestNewWorkerPod_NoSidecar(t *testing.T) {
	build := &v1.Secret{}
	proj := &v1.Secret{
		Data: map[string][]byte{
			"vcsSidecar": []byte(""),
		},
	}
	config := &Config{
		Namespace: v1.NamespaceDefault,
	}

	pod := NewWorkerPod(build, proj, config)

	spec := pod.Spec
	sidecarVolumeExists := false
	for _, volume := range spec.Volumes {
		if volume.Name == "vcs-sidecar" {
			sidecarVolumeExists = true
		}
	}
	if sidecarVolumeExists {
		t.Error("expected vcs-sidecar volume not to exist")
	}
	if len(spec.InitContainers) > 0 {
		t.Error("expected spec.InitContainers to be empty")
	}

	container := spec.Containers[0]

	sidecarVolumeMountExists := false
	for _, volumeMount := range container.VolumeMounts {
		if volumeMount.Name == "vcs-sidecar" {
			sidecarVolumeMountExists = true
		}
	}
	if sidecarVolumeMountExists {
		t.Error("expected vcs-sidecar volume mount not to exist")
	}
}

func TestNewWorkerPod_WorkerEnv_ServiceAccount(t *testing.T) {
	testcases := []struct {
		name        string
		data        map[string][]byte
		config      Config
		wantSA      string
		wantSARegex string
	}{
		{"defaults",
			map[string][]byte{},
			Config{
				ProjectServiceAccount:      DefaultJobServiceAccountName,
				ProjectServiceAccountRegex: DefaultJobServiceAccountName,
			},
			DefaultJobServiceAccountName,
			DefaultJobServiceAccountName,
		},
		{"service account override",
			map[string][]byte{
				"serviceAccount": []byte("my-serviceaccount"),
			},
			Config{
				ProjectServiceAccount:      DefaultJobServiceAccountName,
				ProjectServiceAccountRegex: DefaultJobServiceAccountName,
			},
			"my-serviceaccount",
			"my-serviceaccount",
		},
		{"custom service account regex retained",
			map[string][]byte{
				"serviceAccount": []byte("my-serviceaccount"),
			},
			Config{
				ProjectServiceAccount:      DefaultJobServiceAccountName,
				ProjectServiceAccountRegex: "my-custom-regex-*",
			},
			"my-serviceaccount",
			"my-custom-regex-*",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			build := &v1.Secret{}
			proj := &v1.Secret{
				Data: tc.data,
			}
			config := &tc.config

			pod := NewWorkerPod(build, proj, config)

			spec := pod.Spec
			saEnvFound := false
			for _, env := range spec.Containers[0].Env {
				if env.Name == "BRIGADE_SERVICE_ACCOUNT" {
					if env.Value != tc.wantSA {
						t.Error("Pod has incorrect value for environment variable BRIGADE_SERVICE_ACCOUNT")
					}
					saEnvFound = true
				}
			}

			if !saEnvFound {
				t.Error("Pod is missing environment variable BRIGADE_SERVICE_ACCOUNT")
			}

			// The service account regex value is also expected to be set
			regexEnvFound := false
			for _, env := range spec.Containers[0].Env {
				if env.Name == "BRIGADE_SERVICE_ACCOUNT_REGEX" {
					if env.Value != tc.wantSARegex {
						t.Error("Pod has incorrect value for environment variable BRIGADE_SERVICE_ACCOUNT_REGEX")
					}
					regexEnvFound = true
				}
			}

			if !regexEnvFound {
				t.Error("Pod is missing environment variable BRIGADE_SERVICE_ACCOUNT_REGEX")
			}
		})
	}
}

func TestNewWorkerPod_Affinity(t *testing.T) {
	testcases := []struct {
		name   string
		data   map[string][]byte
		config Config
	}{
		{"without affinity",
			map[string][]byte{},
			Config{
				WorkerNodePoolKey:   "",
				WorkerNodePoolValue: "",
			},
		},
		{"working affinity",
			map[string][]byte{},
			Config{
				WorkerNodePoolKey:   "nodepool-key",
				WorkerNodePoolValue: "nodepool-value",
			},
		},
		{"broken affinity - no nodepool-key",
			map[string][]byte{},
			Config{
				WorkerNodePoolKey:   "",
				WorkerNodePoolValue: "nodepool-value",
			},
		},
		{"broken affinity - no nodepool-value",
			map[string][]byte{},
			Config{
				WorkerNodePoolKey:   "nodepool-key",
				WorkerNodePoolValue: "",
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Running %s", tc.name)
			build := &v1.Secret{}
			proj := &v1.Secret{
				Data: tc.data,
			}
			config := &tc.config

			pod := NewWorkerPod(build, proj, config)

			spec := pod.Spec
			for _, env := range spec.Containers[0].Env {

				if env.Name == "BRIGADE_WORKER_NODEPOOL_KEY" && config.WorkerNodePoolKey == "" {
					t.Error("Env variable BRIGADE_WORKER_NODEPOOL_KEY should not be present")
				}

				if env.Name == "BRIGADE_WORKER_NODEPOOL_VALUE" && config.WorkerNodePoolValue == "" {
					t.Error("Env variable BRIGADE_WORKER_NODEPOOL_VALUE should not be present")
				}
			}
			switch pod.Spec.Affinity {
			case nil:
				if config.WorkerNodePoolKey != "" && config.WorkerNodePoolValue != "" {
					t.Error("Affinity is not defined, while WorkerNodePoolKey and WorkerNodePoolValue are present")
				}
			default:
				if pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0].Key != config.WorkerNodePoolKey {
					t.Errorf("Broken affinity pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0].Key not equal to WorkerNodePoolKey")
				}
				if pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0].Values[0] != config.WorkerNodePoolValue {
					t.Errorf("Broken affinity no pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0].Values[0] not equal WorkerNodePoolValue")
				}
			}

		})
	}
}
