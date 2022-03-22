package kubernetes

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/brigadecore/brigade/v2/apiserver/internal/api"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

// secretsStore is a Kubernetes-based implementation of the api.SecretsStore
// interface.
type secretsStore struct {
	kubeClient kubernetes.Interface
}

// NewSecretsStore returns a Kubernetes-based implementation of the
// api.SecretsStore interface.
func NewSecretsStore(kubeClient kubernetes.Interface) api.SecretsStore {
	return &secretsStore{
		kubeClient: kubeClient,
	}
}

func (s *secretsStore) List(
	ctx context.Context,
	project api.Project,
	opts meta.ListOptions,
) (meta.List[api.Secret], error) {
	secrets := meta.List[api.Secret]{}

	k8sSecret, err := s.kubeClient.CoreV1().Secrets(
		project.Kubernetes.Namespace,
	).Get(ctx, "project-secrets", metav1.GetOptions{})
	if err != nil {
		return secrets, errors.Wrapf(
			err,
			"error retrieving secret \"project-secrets\" in namespace %q",
			project.Kubernetes.Namespace,
		)
	}
	secrets.Items = make([]api.Secret, len(k8sSecret.Data))
	var i int
	for key := range k8sSecret.Data {
		secrets.Items[i] = api.Secret{Key: key}
		i++
	}

	secrets.Sort(func(lhs, rhs api.Secret) int {
		if lhs.Key < rhs.Key {
			return -1
		}
		if lhs.Key == rhs.Key {
			return 0
		}
		return 1
	})

	// Paginate...

	// Technically, it's really kind of pointless to do this. The main reason we
	// paginate any sort of response is to avoid causing OOMs by reading gigantic
	// collections (like millions of Events) into memory, but here, all of these
	// secrets are ALREADY in memory, so we're not ready avoiding any real problem
	// here. But we're going to do it anyway just for the sake of making the
	// ListSecrets operation behave consistently with all other list operations.
	if opts.Continue != "" {
		for i := int64(0); i < secrets.Len(); i++ {
			if secrets.Items[i].Key == opts.Continue {
				secrets.Items = secrets.Items[i+1:]
				break
			}
		}
	}
	if secrets.Len() > opts.Limit {
		secrets.RemainingItemCount = secrets.Len() - opts.Limit
		secrets.Items = secrets.Items[:opts.Limit]
		secrets.Continue = secrets.Items[opts.Limit-1].Key
	}

	return secrets, nil
}

func (s *secretsStore) Set(
	ctx context.Context,
	project api.Project,
	secret api.Secret,
) error {
	patch := struct {
		Data map[string]string `json:"data"`
	}{
		Data: map[string]string{
			secret.Key: base64.StdEncoding.EncodeToString([]byte(secret.Value)),
		},
	}
	patchBytes, err := json.Marshal(patch)
	if err != nil {
		return errors.Wrapf(
			err,
			"error marshaling patch for project %q secret in namespace %q",
			project.ID,
			project.Kubernetes.Namespace,
		)
	}
	if _, err := s.kubeClient.CoreV1().Secrets(
		project.Kubernetes.Namespace,
	).Patch(
		ctx,
		"project-secrets",
		types.StrategicMergePatchType,
		patchBytes,
		metav1.PatchOptions{},
	); err != nil {
		return errors.Wrapf(
			err,
			"error patching project %q secret in namespace %q",
			project.ID,
			project.Kubernetes.Namespace,
		)
	}
	return nil
}

func (s *secretsStore) Unset(
	ctx context.Context,
	project api.Project,
	key string,
) error {
	// Note: If we blindly try to patch the k8s secret to remove the specified
	// key, we'll get an error if that key isn't in the map, so we retrieve the
	// k8s secret and have a peek first. If that key is undefined, we bail early
	// and return no error.
	k8sSecret, err := s.kubeClient.CoreV1().Secrets(
		project.Kubernetes.Namespace,
	).Get(ctx, "project-secrets", metav1.GetOptions{})
	if err != nil {
		return errors.Wrapf(
			err,
			"error retrieving project %q secret in namespace %q",
			project.ID,
			project.Kubernetes.Namespace,
		)
	}
	if _, ok := k8sSecret.Data[key]; !ok {
		return nil
	}
	patch := []struct {
		Op   string `json:"op"`
		Path string `json:"path"`
	}{
		{
			Op:   "remove",
			Path: fmt.Sprintf("/data/%s", key),
		},
	}
	patchBytes, err := json.Marshal(patch)
	if err != nil {
		return errors.Wrapf(
			err,
			"error marshaling patch for project %q secret in namespace %q",
			project.ID,
			project.Kubernetes.Namespace,
		)
	}
	if _, err := s.kubeClient.CoreV1().Secrets(
		project.Kubernetes.Namespace,
	).Patch(
		ctx,
		"project-secrets",
		types.JSONPatchType,
		patchBytes,
		metav1.PatchOptions{},
	); err != nil {
		return errors.Wrapf(
			err,
			"error patching project %q secret in namespace %q",
			project.ID,
			project.Kubernetes.Namespace,
		)
	}
	return nil
}
