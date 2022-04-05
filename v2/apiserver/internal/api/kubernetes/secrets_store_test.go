package kubernetes

import (
	"context"
	"testing"

	"github.com/brigadecore/brigade/v2/apiserver/internal/api"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestNewSecretsStore(t *testing.T) {
	kubeClient := fake.NewSimpleClientset()
	s, ok := NewSecretsStore(kubeClient).(*secretsStore)
	require.True(t, ok)
	require.Same(t, kubeClient, s.kubeClient)
}

func TestSecretsStoreList(t *testing.T) {
	const testNamespace = "foo"
	const testLimit = 3
	const testFirstKey = "abc"
	testCases := []struct {
		name       string
		setup      func() *fake.Clientset
		assertions func(meta.List[api.Secret], error)
	}{
		{
			name: "error getting kubernetes secret",
			setup: func() *fake.Clientset {
				// We'll force an error simply by having the secret not exist
				return fake.NewSimpleClientset()
			},
			assertions: func(secrets meta.List[api.Secret], err error,
			) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "error retrieving secret")
			},
		},

		{
			name: "success",
			setup: func() *fake.Clientset {
				kubeClient := fake.NewSimpleClientset()
				_, err := kubeClient.CoreV1().Secrets(testNamespace).Create(
					context.Background(),
					&corev1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name: "project-secrets",
						},
						// These keys are deliberately out of order, because we will want
						// to test that the order is corrected when they are retrieved.
						//
						// The values below also aren't base64 encoded and that's perfectly
						// fine. The fake.Clientset is really dumb. It just takes what you
						// give it and returns what you give it.
						Data: map[string][]byte{
							"xyz":        []byte("789"),
							"foo":        []byte("bar"),
							"bat":        []byte("baz"),
							testFirstKey: []byte("123"),
						},
					},
					metav1.CreateOptions{},
				)
				require.NoError(t, err)
				return kubeClient
			},
			assertions: func(secrets meta.List[api.Secret], err error) {
				require.NoError(t, err)
				// Check that the Limit param was respected
				require.Len(t, secrets.Items, testLimit)
				// Check that we got Secrets back, lexically ordered by Key AND the
				// Continue param was respected
				require.Equal(t, "bat", secrets.Items[0].Key)
				require.Equal(t, "foo", secrets.Items[1].Key)
				require.Equal(t, "xyz", secrets.Items[2].Key)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			kubeClient := testCase.setup()
			s := &secretsStore{
				kubeClient: kubeClient,
			}
			secrets, err := s.List(
				context.Background(),
				api.Project{
					Kubernetes: &api.KubernetesDetails{
						Namespace: testNamespace,
					},
				},
				meta.ListOptions{
					Continue: testFirstKey,
					Limit:    testLimit,
				},
			)
			testCase.assertions(secrets, err)
		})
	}
}

func TestSecretsStoreSet(t *testing.T) {
	const testNamespace = "foo"
	const testKey = "foo"
	const testValue = "bar"
	testCases := []struct {
		name       string
		setup      func() *fake.Clientset
		assertions func(error, *fake.Clientset)
	}{
		{
			name: "error getting kubernetes secret",
			setup: func() *fake.Clientset {
				// We'll force an error simply by having the secret not exist
				return fake.NewSimpleClientset()
			},
			assertions: func(err error, kubeClient *fake.Clientset) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "error patching")
			},
		},

		{
			name: "success",
			setup: func() *fake.Clientset {
				kubeClient := fake.NewSimpleClientset()
				_, err := kubeClient.CoreV1().Secrets(testNamespace).Create(
					context.Background(),
					&corev1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name: "project-secrets",
						},
					},
					metav1.CreateOptions{},
				)
				require.NoError(t, err)
				return kubeClient
			},
			assertions: func(err error, kubeClient *fake.Clientset) {
				require.NoError(t, err)
				secret, err := kubeClient.CoreV1().Secrets(testNamespace).Get(
					context.Background(),
					"project-secrets",
					metav1.GetOptions{},
				)
				require.NoError(t, err)
				require.Equal(t, testValue, string(secret.Data[testKey]))
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			kubeClient := testCase.setup()
			s := &secretsStore{
				kubeClient: kubeClient,
			}
			err := s.Set(
				context.Background(),
				api.Project{
					Kubernetes: &api.KubernetesDetails{
						Namespace: testNamespace,
					},
				},
				api.Secret{Key: testKey, Value: testValue},
			)
			testCase.assertions(err, kubeClient)
		})
	}
}

func TestSecretsStoreUnset(t *testing.T) {
	const testNamespace = "foo"
	const testKey = "foo"
	const testValue = "bar"
	testCases := []struct {
		name       string
		setup      func() *fake.Clientset
		assertions func(error, *fake.Clientset)
	}{
		{
			name: "error getting kubernetes secret",
			setup: func() *fake.Clientset {
				// We'll force an error simply by having the secret not exist
				return fake.NewSimpleClientset()
			},
			assertions: func(err error, kubeClient *fake.Clientset) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "error retrieving project")
			},
		},

		{
			name: "key doesn't exist in kubernetes secret",
			setup: func() *fake.Clientset {
				kubeClient := fake.NewSimpleClientset()
				_, err := kubeClient.CoreV1().Secrets(testNamespace).Create(
					context.Background(),
					&corev1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name: "project-secrets",
						},
					},
					metav1.CreateOptions{},
				)
				require.NoError(t, err)
				return kubeClient
			},
			assertions: func(err error, _ *fake.Clientset) {
				require.NoError(t, err)
			},
		},

		{
			name: "success",
			setup: func() *fake.Clientset {
				kubeClient := fake.NewSimpleClientset()
				_, err := kubeClient.CoreV1().Secrets(testNamespace).Create(
					context.Background(),
					&corev1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name: "project-secrets",
						},
						// The values below also aren't base64 encoded and that's perfectly
						// fine. The fake.Clientset is really dumb. It just takes what you
						// give it and returns what you give it.
						Data: map[string][]byte{
							testKey: []byte(testValue),
						},
					},
					metav1.CreateOptions{},
				)
				require.NoError(t, err)
				return kubeClient
			},
			assertions: func(err error, kubeClient *fake.Clientset) {
				require.NoError(t, err)
				secret, err := kubeClient.CoreV1().Secrets(testNamespace).Get(
					context.Background(),
					"project-secrets",
					metav1.GetOptions{},
				)
				require.NoError(t, err)
				_, ok := secret.Data[testKey]
				require.False(t, ok)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			kubeClient := testCase.setup()
			s := &secretsStore{
				kubeClient: kubeClient,
			}
			err := s.Unset(
				context.Background(),
				api.Project{
					Kubernetes: &api.KubernetesDetails{
						Namespace: testNamespace,
					},
				},
				testKey,
			)
			testCase.assertions(err, kubeClient)
		})
	}
}
