package kubernetes

import (
	"context"
	"strings"
	"testing"

	"github.com/brigadecore/brigade/v2/apiserver/internal/core"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestNewSubstrate(t *testing.T) {
	testClient := fake.NewSimpleClientset()
	s := NewSubstrate(testClient)
	require.IsType(t, &substrate{}, s)
	require.Same(t, testClient, s.(*substrate).kubeClient)
}

func TestSubstrateCreateProject(t *testing.T) {
	const testNamespace = "foo"
	testCases := []struct {
		name       string
		setup      func() *fake.Clientset
		assertions func(core.Project, error, *fake.Clientset)
	}{
		{
			name: "error creating namespace",
			setup: func() *fake.Clientset {
				kubeClient := fake.NewSimpleClientset()
				// We'll force an error due to the Namespace already existing
				_, err := kubeClient.CoreV1().Namespaces().Create(
					context.Background(),
					&corev1.Namespace{
						ObjectMeta: metav1.ObjectMeta{
							Name: testNamespace,
						},
					},
					metav1.CreateOptions{},
				)
				require.NoError(t, err)
				return kubeClient
			},
			assertions: func(
				project core.Project,
				err error,
				kubeClient *fake.Clientset,
			) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "error creating namespace")
				require.Contains(t, err.Error(), "already exists")
			},
		},

		{
			name: "error creating workers role",
			setup: func() *fake.Clientset {
				kubeClient := fake.NewSimpleClientset()
				// We'll force an error due to the Role already existing
				_, err := kubeClient.RbacV1().Roles(testNamespace).Create(
					context.Background(),
					&rbacv1.Role{
						ObjectMeta: metav1.ObjectMeta{
							Name: "workers",
						},
					},
					metav1.CreateOptions{},
				)
				require.NoError(t, err)
				return kubeClient
			},
			assertions: func(
				project core.Project,
				err error,
				kubeClient *fake.Clientset,
			) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "error creating role")
				require.Contains(t, err.Error(), "already exists")
			},
		},

		{
			name: "error creating workers service account",
			setup: func() *fake.Clientset {
				kubeClient := fake.NewSimpleClientset()
				// We'll force an error due to the ServiceAccount already existing
				_, err := kubeClient.CoreV1().ServiceAccounts(testNamespace).Create(
					context.Background(),
					&corev1.ServiceAccount{
						ObjectMeta: metav1.ObjectMeta{
							Name: "workers",
						},
					},
					metav1.CreateOptions{},
				)
				require.NoError(t, err)
				return kubeClient
			},
			assertions: func(
				project core.Project,
				err error,
				kubeClient *fake.Clientset,
			) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "error creating service account")
				require.Contains(t, err.Error(), "already exists")
			},
		},

		{
			name: "error creating workers role binding",
			setup: func() *fake.Clientset {
				kubeClient := fake.NewSimpleClientset()
				// We'll force an error due to the RoleBinding already existing
				_, err := kubeClient.RbacV1().RoleBindings(testNamespace).Create(
					context.Background(),
					&rbacv1.RoleBinding{
						ObjectMeta: metav1.ObjectMeta{
							Name: "workers",
						},
					},
					metav1.CreateOptions{},
				)
				require.NoError(t, err)
				return kubeClient
			},
			assertions: func(
				project core.Project,
				err error,
				kubeClient *fake.Clientset,
			) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "error creating role binding")
				require.Contains(t, err.Error(), "already exists")
			},
		},

		{
			name: "error creating jobs role",
			setup: func() *fake.Clientset {
				kubeClient := fake.NewSimpleClientset()
				// We'll force an error due to the Role already existing
				_, err := kubeClient.RbacV1().Roles(testNamespace).Create(
					context.Background(),
					&rbacv1.Role{
						ObjectMeta: metav1.ObjectMeta{
							Name: "jobs",
						},
					},
					metav1.CreateOptions{},
				)
				require.NoError(t, err)
				return kubeClient
			},
			assertions: func(
				project core.Project,
				err error,
				kubeClient *fake.Clientset,
			) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "error creating role")
				require.Contains(t, err.Error(), "already exists")
			},
		},

		{
			name: "error creating jobs service account",
			setup: func() *fake.Clientset {
				kubeClient := fake.NewSimpleClientset()
				// We'll force an error due to the ServiceAccount already existing
				_, err := kubeClient.CoreV1().ServiceAccounts(testNamespace).Create(
					context.Background(),
					&corev1.ServiceAccount{
						ObjectMeta: metav1.ObjectMeta{
							Name: "jobs",
						},
					},
					metav1.CreateOptions{},
				)
				require.NoError(t, err)
				return kubeClient
			},
			assertions: func(
				project core.Project,
				err error,
				kubeClient *fake.Clientset,
			) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "error creating service account")
				require.Contains(t, err.Error(), "already exists")
			},
		},

		{
			name: "error creating jobs role binding",
			setup: func() *fake.Clientset {
				kubeClient := fake.NewSimpleClientset()
				// We'll force an error due to the RoleBinding already existing
				_, err := kubeClient.RbacV1().RoleBindings(testNamespace).Create(
					context.Background(),
					&rbacv1.RoleBinding{
						ObjectMeta: metav1.ObjectMeta{
							Name: "jobs",
						},
					},
					metav1.CreateOptions{},
				)
				require.NoError(t, err)
				return kubeClient
			},
			assertions: func(
				project core.Project,
				err error,
				kubeClient *fake.Clientset,
			) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "error creating role binding")
				require.Contains(t, err.Error(), "already exists")
			},
		},

		{
			name: "error creating project secret",
			setup: func() *fake.Clientset {
				kubeClient := fake.NewSimpleClientset()
				// We'll force an error due to the Secret already existing
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
			assertions: func(
				project core.Project,
				err error,
				kubeClient *fake.Clientset,
			) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "error creating secret")
				require.Contains(t, err.Error(), "already exists")
			},
		},

		{
			name: "success",
			setup: func() *fake.Clientset {
				return fake.NewSimpleClientset()
			},
			assertions: func(
				project core.Project,
				err error,
				kubeClient *fake.Clientset,
			) {
				require.NoError(t, err)

				// Check that the project was augmented with Kubernetes-specific details
				require.NotNil(t, project.Kubernetes)
				require.NotEmpty(t, project.Kubernetes.Namespace)

				// Check that an RBAC Role was created for the Project's Workers
				role, err := kubeClient.RbacV1().Roles(
					project.Kubernetes.Namespace,
				).Get(context.Background(), "workers", v1.GetOptions{})
				require.NoError(t, err)
				require.NotNil(t, role)

				// Check that a ServiceAccount was created for the Project's Workers
				servicAccount, err := kubeClient.CoreV1().ServiceAccounts(
					project.Kubernetes.Namespace,
				).Get(context.Background(), "workers", v1.GetOptions{})
				require.NoError(t, err)
				require.NotNil(t, servicAccount)

				// Check that an RBAC RoleBinding associates the Workers' ServiceAccount
				// with the Workers' RBAC Role
				roleBinding, err := kubeClient.RbacV1().RoleBindings(
					project.Kubernetes.Namespace,
				).Get(context.Background(), "workers", v1.GetOptions{})
				require.NoError(t, err)
				require.NotNil(t, roleBinding)

				// Check that an RBAC Role was created for the Project's Jobs
				role, err = kubeClient.RbacV1().Roles(
					project.Kubernetes.Namespace,
				).Get(context.Background(), "jobs", v1.GetOptions{})
				require.NoError(t, err)
				require.NotNil(t, role)

				// Check that a ServiceAccount was created for the Project's Jobs
				servicAccount, err = kubeClient.CoreV1().ServiceAccounts(
					project.Kubernetes.Namespace,
				).Get(context.Background(), "jobs", v1.GetOptions{})
				require.NoError(t, err)
				require.NotNil(t, servicAccount)

				// Check that an RBAC RoleBinding associates the Jobs' ServiceAccount
				// with the Jobs' RBAC Role
				roleBinding, err = kubeClient.RbacV1().RoleBindings(
					project.Kubernetes.Namespace,
				).Get(context.Background(), "jobs", v1.GetOptions{})
				require.NoError(t, err)
				require.NotNil(t, roleBinding)

				// Check that a Secret was created to store the Project's Secrets
				secrets, err := kubeClient.CoreV1().Secrets(
					project.Kubernetes.Namespace,
				).Get(context.Background(), "project-secrets", v1.GetOptions{})
				require.NoError(t, err)
				require.NotNil(t, secrets)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			kubeClient := testCase.setup()
			s := &substrate{
				generateNewNamespaceFn: func(string) string {
					return testNamespace
				},
				kubeClient: kubeClient,
			}
			project, err := s.CreateProject(context.Background(), core.Project{})
			testCase.assertions(project, err, kubeClient)
		})
	}
}

func TestSubstrateDeleteProject(t *testing.T) {
	const testNamespace = "foo"
	testCases := []struct {
		name       string
		setup      func() *fake.Clientset
		assertions func(error, *fake.Clientset)
	}{
		{
			name: "error deleting namespace",
			setup: func() *fake.Clientset {
				return fake.NewSimpleClientset()
			},
			assertions: func(err error, kubeClient *fake.Clientset) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "not found")
				require.Contains(t, err.Error(), "error deleting namespace")
			},
		},

		{
			name: "success",
			setup: func() *fake.Clientset {
				kubeClient := fake.NewSimpleClientset()
				// Make sure the Namespace exists so it can be deleted
				_, err := kubeClient.CoreV1().Namespaces().Create(
					context.Background(),
					&corev1.Namespace{
						ObjectMeta: metav1.ObjectMeta{
							Name: testNamespace,
						},
					},
					metav1.CreateOptions{},
				)
				require.NoError(t, err)
				return kubeClient
			},
			assertions: func(err error, kubeClient *fake.Clientset) {
				require.NoError(t, err)

				// Check that the Namespace is gone
				_, err = kubeClient.CoreV1().Namespaces().Get(
					context.Background(),
					testNamespace,
					v1.GetOptions{},
				)
				require.Error(t, err)
				require.Contains(t, err.Error(), "not found")
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			kubeClient := testCase.setup()
			s := &substrate{
				generateNewNamespaceFn: func(string) string {
					return testNamespace
				},
				kubeClient: kubeClient,
			}
			err := s.DeleteProject(
				context.Background(), core.Project{
					Kubernetes: &core.KubernetesDetails{
						Namespace: testNamespace,
					},
				},
			)
			testCase.assertions(err, kubeClient)
		})
	}
}

// TODO: Find a better way to test this. Unfortunately, the DeleteCollection
// function on a *fake.ClientSet doesn't ACTUALLY delete collections of
// resources based on the labels provided.
//
// Refer to: https://github.com/kubernetes/client-go/issues/609
//
// This makes it basically impossible to assert what we'd LIKE to assert here--
// that resources labeled with the correct Event ID are deleted while other
// resources are left alone. We'll settle for invoking DeleteWorkerAndJobs(...)
// and asserting we get no error-- so we at least get some test coverage for
// this function. We'll have to make sure this behavior is well-covered by
// integration or e2e tests in the future.
func TestSubstrateDeleteWorkerAndJobs(t *testing.T) {
	s := &substrate{
		kubeClient: fake.NewSimpleClientset(),
	}
	err := s.DeleteWorkerAndJobs(
		context.Background(),
		core.Project{
			Kubernetes: &core.KubernetesDetails{
				Namespace: "foo",
			},
		},
		core.Event{
			ObjectMeta: meta.ObjectMeta{
				ID: "bar",
			},
		},
	)
	require.NoError(t, err)
}

func TestGenerateNewNamespace(t *testing.T) {
	const testProjectID = "foo"
	namespace := generateNewNamespace(testProjectID)
	tokens := strings.SplitN(namespace, "-", 3)
	require.Len(t, tokens, 3)
	require.Equal(t, "brigade", tokens[0])
	require.Equal(t, testProjectID, tokens[1])
	_, err := uuid.FromString(tokens[2])
	require.NoError(t, err)
}
