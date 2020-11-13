package kubernetes

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/brigadecore/brigade/v2/apiserver/internal/core"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/queue"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	myk8s "github.com/brigadecore/brigade/v2/internal/kubernetes"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

func TestNewSubstrate(t *testing.T) {
	testClient := fake.NewSimpleClientset()
	testQueueWriterFactory := &mockQueueWriterFactory{}
	testConfig := SubstrateConfig{}
	s := NewSubstrate(testClient, testQueueWriterFactory, testConfig)
	require.IsType(t, &substrate{}, s)
	require.Same(t, testClient, s.(*substrate).kubeClient)
	require.Same(t, testQueueWriterFactory, s.(*substrate).queueWriterFactory)
	require.Equal(t, testConfig, s.(*substrate).config)
}

func TestSubstrateCountRunningWorkers(t *testing.T) {
	const testNamespace = "foo"
	kubeClient := fake.NewSimpleClientset()
	podsClient := kubeClient.CoreV1().Pods(testNamespace)
	// This pod doesn't have correct labels
	_, err := podsClient.Create(
		context.Background(),
		&corev1.Pod{
			ObjectMeta: v1.ObjectMeta{
				Name: "bar",
				Labels: map[string]string{
					myk8s.LabelComponent: "job",
				},
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodRunning,
			},
		},
		v1.CreateOptions{},
	)
	require.NoError(t, err)
	// This pod has correct labels
	_, err = podsClient.Create(
		context.Background(),
		&corev1.Pod{
			ObjectMeta: v1.ObjectMeta{
				Name: "bat",
				Labels: map[string]string{
					myk8s.LabelComponent: "worker",
				},
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodRunning,
			},
		},
		v1.CreateOptions{},
	)
	require.NoError(t, err)
	s := &substrate{
		kubeClient: kubeClient,
	}
	count, err := s.CountRunningWorkers(context.Background())
	require.NoError(t, err)
	require.Equal(t, 1, count.Count)
}

func TestSubstrateCountRunningJobs(t *testing.T) {
	const testNamespace = "foo"
	kubeClient := fake.NewSimpleClientset()
	podsClient := kubeClient.CoreV1().Pods(testNamespace)
	// This pod doesn't have correct labels
	_, err := podsClient.Create(
		context.Background(),
		&corev1.Pod{
			ObjectMeta: v1.ObjectMeta{
				Name: "bar",
				Labels: map[string]string{
					myk8s.LabelComponent: "worker",
				},
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodRunning,
			},
		},
		v1.CreateOptions{},
	)
	require.NoError(t, err)
	// This pod has correct labels
	_, err = podsClient.Create(
		context.Background(),
		&corev1.Pod{
			ObjectMeta: v1.ObjectMeta{
				Name: "bat",
				Labels: map[string]string{
					myk8s.LabelComponent: "job",
				},
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodRunning,
			},
		},
		v1.CreateOptions{},
	)
	require.NoError(t, err)
	s := &substrate{
		kubeClient: kubeClient,
	}
	count, err := s.CountRunningJobs(context.Background())
	require.NoError(t, err)
	require.Equal(t, 1, count.Count)
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

func TestSubstrateScheduleWorker(t *testing.T) {
	const testNamespace = "foo"
	const testEventID = "12345"
	testCases := []struct {
		name       string
		setup      func() core.Substrate
		assertions func(error)
	}{
		{
			name: "error getting project secret",
			setup: func() core.Substrate {
				return &substrate{
					kubeClient: fake.NewSimpleClientset(),
				}
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(
					t,
					err.Error(),
					"error finding secret \"project-secrets\"",
				)
			},
		},

		{
			name: "error creating event secret",
			setup: func() core.Substrate {
				kubeClient := fake.NewSimpleClientset()
				_, err := kubeClient.CoreV1().Secrets(testNamespace).Create(
					context.Background(),
					&corev1.Secret{
						ObjectMeta: v1.ObjectMeta{
							Name: "project-secrets",
						},
					},
					v1.CreateOptions{},
				)
				require.NoError(t, err)
				// We'll force an error creating the event secret by having it already
				// exist
				_, err = kubeClient.CoreV1().Secrets(testNamespace).Create(
					context.Background(),
					&corev1.Secret{
						ObjectMeta: v1.ObjectMeta{
							Name: fmt.Sprintf("event-%s", testEventID),
						},
					},
					v1.CreateOptions{},
				)
				require.NoError(t, err)
				return &substrate{
					kubeClient: kubeClient,
				}
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "error creating secret")
			},
		},

		{
			name: "error creating queue writer",
			setup: func() core.Substrate {
				kubeClient := fake.NewSimpleClientset()
				_, err := kubeClient.CoreV1().Secrets(testNamespace).Create(
					context.Background(),
					&corev1.Secret{
						ObjectMeta: v1.ObjectMeta{
							Name: "project-secrets",
						},
					},
					v1.CreateOptions{},
				)
				require.NoError(t, err)
				return &substrate{
					kubeClient: kubeClient,
					queueWriterFactory: &mockQueueWriterFactory{
						NewWriterFn: func(queueName string) (queue.Writer, error) {
							return nil, errors.New("something went wrong")
						},
					},
				}
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error creating queue writer")
			},
		},

		{
			name: "error writing to queue",
			setup: func() core.Substrate {
				kubeClient := fake.NewSimpleClientset()
				_, err := kubeClient.CoreV1().Secrets(testNamespace).Create(
					context.Background(),
					&corev1.Secret{
						ObjectMeta: v1.ObjectMeta{
							Name: "project-secrets",
						},
					},
					v1.CreateOptions{},
				)
				require.NoError(t, err)
				return &substrate{
					kubeClient: kubeClient,
					queueWriterFactory: &mockQueueWriterFactory{
						NewWriterFn: func(queueName string) (queue.Writer, error) {
							return &mockQueueWriter{
								WriteFn: func(context.Context, string) error {
									return errors.New("something went wrong")
								},
								CloseFn: func(context.Context) error {
									return nil
								},
							}, nil
						},
					},
				}
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(
					t,
					err.Error(),
					"error submitting execution task for event",
				)
			},
		},

		{
			name: "success",
			setup: func() core.Substrate {
				kubeClient := fake.NewSimpleClientset()
				_, err := kubeClient.CoreV1().Secrets(testNamespace).Create(
					context.Background(),
					&corev1.Secret{
						ObjectMeta: v1.ObjectMeta{
							Name: "project-secrets",
						},
					},
					v1.CreateOptions{},
				)
				require.NoError(t, err)
				return &substrate{
					kubeClient: kubeClient,
					queueWriterFactory: &mockQueueWriterFactory{
						NewWriterFn: func(queueName string) (queue.Writer, error) {
							return &mockQueueWriter{
								WriteFn: func(context.Context, string) error {
									return nil
								},
								CloseFn: func(context.Context) error {
									return nil
								},
							}, nil
						},
					},
				}
			},
			assertions: func(err error) {
				require.NoError(t, err)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			substrate := testCase.setup()
			err := substrate.ScheduleWorker(
				context.Background(),
				core.Project{
					Kubernetes: &core.KubernetesDetails{
						Namespace: testNamespace,
					},
				},
				core.Event{
					ObjectMeta: meta.ObjectMeta{
						ID: testEventID,
					},
				},
			)
			testCase.assertions(err)
		})
	}
}

func TestSubstrateStartWorker(t *testing.T) {
	testCases := []struct {
		name       string
		substrate  core.Substrate
		assertions func(error)
	}{
		{
			name: "error creating workspace",
			substrate: &substrate{
				createWorkspacePVCFn: func(
					context.Context,
					core.Project,
					core.Event,
				) error {
					return errors.New("something went wrong")
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error creating workspace for event")
			},
		},
		{
			name: "error creating worker pod",
			substrate: &substrate{
				createWorkspacePVCFn: func(
					context.Context,
					core.Project,
					core.Event,
				) error {
					return nil
				},
				createWorkerPodFn: func(
					context.Context,
					core.Project,
					core.Event,
				) error {
					return errors.New("something went wrong")
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error creating pod for event")
			},
		},
		{
			name: "success",
			substrate: &substrate{
				createWorkspacePVCFn: func(
					context.Context,
					core.Project,
					core.Event,
				) error {
					return nil
				},
				createWorkerPodFn: func(
					context.Context,
					core.Project,
					core.Event,
				) error {
					return nil
				},
			},
			assertions: func(err error) {
				require.NoError(t, err)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := testCase.substrate.StartWorker(
				context.Background(),
				core.Project{},
				core.Event{
					Worker: core.Worker{
						Spec: core.WorkerSpec{
							UseWorkspace: true,
						},
					},
				},
			)
			testCase.assertions(err)
		})
	}
}

func TestSubstrateStartJob(t *testing.T) {
	const testJobName = "foo"
	testCases := []struct {
		name       string
		substrate  core.Substrate
		assertions func(error)
	}{
		{
			name: "error creating job secret",
			substrate: &substrate{
				createJobSecretFn: func(
					context.Context,
					core.Project,
					core.Event,
					string,
					core.JobSpec,
				) error {
					return errors.New("something went wrong")
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error creating secret for event")
			},
		},
		{
			name: "error creating job pod",
			substrate: &substrate{
				createJobSecretFn: func(
					context.Context,
					core.Project,
					core.Event,
					string,
					core.JobSpec,
				) error {
					return nil
				},
				createJobPodFn: func(
					context.Context,
					core.Project,
					core.Event,
					string,
					core.JobSpec,
				) error {
					return errors.New("something went wrong")
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error creating pod for event")
			},
		},
		{
			name: "success",
			substrate: &substrate{
				createJobSecretFn: func(
					context.Context,
					core.Project,
					core.Event,
					string,
					core.JobSpec,
				) error {
					return nil
				},
				createJobPodFn: func(
					context.Context,
					core.Project,
					core.Event,
					string,
					core.JobSpec,
				) error {
					return nil
				},
			},
			assertions: func(err error) {
				require.NoError(t, err)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := testCase.substrate.StartJob(
				context.Background(),
				core.Project{},
				core.Event{
					Worker: core.Worker{
						Spec: core.WorkerSpec{
							UseWorkspace: true,
						},
					},
				},
				testJobName,
			)
			testCase.assertions(err)
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

func TestSubstrateCreateWorkspacePVC(t *testing.T) {
	testProject := core.Project{
		Kubernetes: &core.KubernetesDetails{
			Namespace: "foo",
		},
	}
	const testEventID = "123456789"
	testCases := []struct {
		name       string
		event      core.Event
		setup      func() *substrate
		assertions func(kubernetes.Interface, error)
	}{
		{
			name: "unparsable storage quantity",
			event: core.Event{
				Worker: core.Worker{
					Spec: core.WorkerSpec{
						WorkspaceSize: "10ZillionBytes",
					},
				},
			},
			setup: func() *substrate {
				return &substrate{}
			},
			assertions: func(_ kubernetes.Interface, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "error parsing storage quantity")
			},
		},
		{
			name: "error creating pvc",
			event: core.Event{
				ObjectMeta: meta.ObjectMeta{
					ID: testEventID,
				},
			},
			setup: func() *substrate {
				kubeClient := fake.NewSimpleClientset()
				// Ensure a failure by pre-creating a PVC with the expected name
				_, err := kubeClient.CoreV1().PersistentVolumeClaims(
					testProject.Kubernetes.Namespace,
				).Create(
					context.Background(),
					&corev1.PersistentVolumeClaim{
						ObjectMeta: v1.ObjectMeta{
							Name: fmt.Sprintf("workspace-%s", testEventID),
						},
					},
					v1.CreateOptions{},
				)
				require.NoError(t, err)
				return &substrate{
					kubeClient: kubeClient,
				}
			},
			assertions: func(_ kubernetes.Interface, err error) {
				require.Error(t, err)
				require.Contains(
					t,
					err.Error(),
					"error creating workspace PVC for event",
				)
			},
		},
		{
			name: "success",
			event: core.Event{
				ObjectMeta: meta.ObjectMeta{
					ID: testEventID,
				},
			},
			setup: func() *substrate {
				return &substrate{
					kubeClient: fake.NewSimpleClientset(),
				}
			},
			assertions: func(kubeClient kubernetes.Interface, err error) {
				require.NoError(t, err)
				pvc, err := kubeClient.CoreV1().PersistentVolumeClaims(
					testProject.Kubernetes.Namespace,
				).Get(
					context.Background(),
					fmt.Sprintf("workspace-%s", testEventID),
					v1.GetOptions{},
				)
				require.NoError(t, err)
				require.NotNil(t, pvc)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			substrate := testCase.setup()
			err := substrate.createWorkspacePVC(
				context.Background(),
				testProject,
				testCase.event,
			)
			testCase.assertions(substrate.kubeClient, err)
		})
	}
}

func TestSubstrateCreateWorkerPod(t *testing.T) {
	testProject := core.Project{
		Kubernetes: &core.KubernetesDetails{
			Namespace: "foo",
		},
	}
	testEvent := core.Event{
		ObjectMeta: meta.ObjectMeta{
			ID: "123456789",
		},
		Worker: core.Worker{
			Spec: core.WorkerSpec{
				Kubernetes: &core.KubernetesConfig{
					ImagePullSecrets: []string{"foo", "bar"},
				},
				UseWorkspace: true,
				Git: &core.GitConfig{
					CloneURL: "a fake clone url",
				},
				Container: &core.ContainerSpec{
					Environment: map[string]string{
						"FOO": "bar",
					},
				},
			},
		},
	}
	testCases := []struct {
		name       string
		setup      func() *substrate
		assertions func(kubernetes.Interface, error)
	}{
		{
			name: "error creating pod",
			setup: func() *substrate {
				kubeClient := fake.NewSimpleClientset()
				// Ensure a failure by pre-creating a pod with the expected name
				_, err := kubeClient.CoreV1().Pods(
					testProject.Kubernetes.Namespace,
				).Create(
					context.Background(),
					&corev1.Pod{
						ObjectMeta: v1.ObjectMeta{
							Name: fmt.Sprintf("worker-%s", testEvent.ID),
						},
					},
					v1.CreateOptions{},
				)
				require.NoError(t, err)
				return &substrate{
					kubeClient: kubeClient,
				}
			},
			assertions: func(_ kubernetes.Interface, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "error creating pod for event")
			},
		},
		{
			name: "success",
			setup: func() *substrate {
				return &substrate{
					kubeClient: fake.NewSimpleClientset(),
				}
			},
			assertions: func(kubeClient kubernetes.Interface, err error) {
				require.NoError(t, err)
				pod, err := kubeClient.CoreV1().Pods(
					testProject.Kubernetes.Namespace,
				).Get(
					context.Background(),
					fmt.Sprintf("worker-%s", testEvent.ID),
					v1.GetOptions{},
				)
				require.NoError(t, err)
				require.NotNil(t, pod)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			substrate := testCase.setup()
			err := substrate.createWorkerPod(
				context.Background(),
				testProject,
				testEvent,
			)
			testCase.assertions(substrate.kubeClient, err)
		})
	}
}

func TestSubstrateCreateJobSecret(t *testing.T) {
	testProject := core.Project{
		Kubernetes: &core.KubernetesDetails{
			Namespace: "foo",
		},
	}
	testEvent := core.Event{
		ObjectMeta: meta.ObjectMeta{
			ID: "123456789",
		},
	}
	const testJobName = "italian"
	testJobSpec := core.JobSpec{
		PrimaryContainer: core.JobContainerSpec{
			ContainerSpec: core.ContainerSpec{
				Environment: map[string]string{
					"FOO": "bar",
				},
			},
		},
		SidecarContainers: map[string]core.JobContainerSpec{
			"helper": {
				ContainerSpec: core.ContainerSpec{
					Environment: map[string]string{
						"BAT": "baz",
					},
				},
			},
		},
	}
	testCases := []struct {
		name       string
		setup      func() *substrate
		assertions func(kubernetes.Interface, error)
	}{
		{
			name: "error creating secret",
			setup: func() *substrate {
				kubeClient := fake.NewSimpleClientset()
				// Ensure a failure by pre-creating a secret with the expected name
				_, err := kubeClient.CoreV1().Secrets(
					testProject.Kubernetes.Namespace,
				).Create(
					context.Background(),
					&corev1.Secret{
						ObjectMeta: v1.ObjectMeta{
							Name: fmt.Sprintf("job-%s-%s", testEvent.ID, testJobName),
						},
					},
					v1.CreateOptions{},
				)
				require.NoError(t, err)
				return &substrate{
					kubeClient: kubeClient,
				}
			},
			assertions: func(_ kubernetes.Interface, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "error creating secret for event")
			},
		},
		{
			name: "success",
			setup: func() *substrate {
				return &substrate{
					kubeClient: fake.NewSimpleClientset(),
				}
			},
			assertions: func(kubeClient kubernetes.Interface, err error) {
				require.NoError(t, err)
				secret, err := kubeClient.CoreV1().Secrets(
					testProject.Kubernetes.Namespace,
				).Get(
					context.Background(),
					fmt.Sprintf("job-%s-%s", testEvent.ID, testJobName),
					v1.GetOptions{},
				)
				require.NoError(t, err)
				require.NotNil(t, secret)
				val, ok := secret.StringData["italian.FOO"]
				require.True(t, ok)
				require.Equal(t, "bar", string(val))
				val, ok = secret.StringData["helper.BAT"]
				require.True(t, ok)
				require.Equal(t, "baz", string(val))
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			substrate := testCase.setup()
			err := substrate.createJobSecret(
				context.Background(),
				testProject,
				testEvent,
				testJobName,
				testJobSpec,
			)
			testCase.assertions(substrate.kubeClient, err)
		})
	}
}

func TestSubstrateCreateJobPod(t *testing.T) {
	testProject := core.Project{
		Kubernetes: &core.KubernetesDetails{
			Namespace: "foo",
		},
	}
	testEvent := core.Event{
		ObjectMeta: meta.ObjectMeta{
			ID: "123456789",
		},
		Worker: core.Worker{
			Spec: core.WorkerSpec{
				Git: &core.GitConfig{
					CloneURL: "a fake git repo url",
				},
				Kubernetes: &core.KubernetesConfig{
					ImagePullSecrets: []string{"foo", "bar"},
				},
			},
		},
	}
	const testJobName = "italian"
	testJobSpec := core.JobSpec{
		PrimaryContainer: core.JobContainerSpec{
			ContainerSpec: core.ContainerSpec{
				Environment: map[string]string{
					"FOO": "bar",
				},
			},
			UseWorkspace:        true,
			UseSource:           true,
			UseHostDockerSocket: true,
			Privileged:          true,
		},
		SidecarContainers: map[string]core.JobContainerSpec{
			"helper": {
				ContainerSpec: core.ContainerSpec{
					Environment: map[string]string{
						"BAT": "baz",
					},
				},
				UseWorkspace:        true,
				UseSource:           true,
				UseHostDockerSocket: true,
				Privileged:          true,
			},
		},
	}
	testCases := []struct {
		name       string
		setup      func() *substrate
		assertions func(kubernetes.Interface, error)
	}{
		{
			name: "error creating pod",
			setup: func() *substrate {
				kubeClient := fake.NewSimpleClientset()
				// Ensure a failure by pre-creating a pod with the expected name
				_, err := kubeClient.CoreV1().Pods(
					testProject.Kubernetes.Namespace,
				).Create(
					context.Background(),
					&corev1.Pod{
						ObjectMeta: v1.ObjectMeta{
							Name: fmt.Sprintf("job-%s-%s", testEvent.ID, testJobName),
						},
					},
					v1.CreateOptions{},
				)
				require.NoError(t, err)
				return &substrate{
					kubeClient: kubeClient,
				}
			},
			assertions: func(_ kubernetes.Interface, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "error creating pod for event")
			},
		},
		{
			name: "success",
			setup: func() *substrate {
				return &substrate{
					kubeClient: fake.NewSimpleClientset(),
				}
			},
			assertions: func(kubeClient kubernetes.Interface, err error) {
				require.NoError(t, err)
				pod, err := kubeClient.CoreV1().Pods(
					testProject.Kubernetes.Namespace,
				).Get(
					context.Background(),
					fmt.Sprintf("job-%s-%s", testEvent.ID, testJobName),
					v1.GetOptions{},
				)
				require.NoError(t, err)
				require.NotNil(t, pod)
				// Volumes:
				require.Len(t, pod.Spec.Volumes, 3)
				require.Equal(t, "workspace", pod.Spec.Volumes[0].Name)
				require.Equal(t, "vcs", pod.Spec.Volumes[1].Name)
				require.Equal(t, "docker-socket", pod.Spec.Volumes[2].Name)
				// Init container:
				require.Len(t, pod.Spec.InitContainers, 1)
				require.Equal(t, "vcs", pod.Spec.InitContainers[0].Name)
				require.Len(t, pod.Spec.InitContainers[0].VolumeMounts, 1)
				require.Equal(t, "vcs", pod.Spec.InitContainers[0].VolumeMounts[0].Name)
				// Containers:
				require.Len(t, pod.Spec.Containers, 2)
				// Primary container:
				require.Equal(t, testJobName, pod.Spec.Containers[0].Name)
				require.Len(t, pod.Spec.Containers[0].Env, 1)
				require.Equal(t, "FOO", pod.Spec.Containers[0].Env[0].Name)
				require.Len(t, pod.Spec.Containers[0].VolumeMounts, 3)
				require.Equal(
					t,
					"workspace",
					pod.Spec.Containers[0].VolumeMounts[0].Name,
				)
				require.Equal(t, "vcs", pod.Spec.Containers[0].VolumeMounts[1].Name)
				require.Equal(
					t,
					"docker-socket",
					pod.Spec.Containers[0].VolumeMounts[2].Name,
				)
				// Sidecar container:
				require.Equal(t, "helper", pod.Spec.Containers[1].Name)
				require.Len(t, pod.Spec.Containers[1].Env, 1)
				require.Equal(t, "BAT", pod.Spec.Containers[1].Env[0].Name)
				require.Len(t, pod.Spec.Containers[1].VolumeMounts, 3)
				require.Equal(
					t,
					"workspace",
					pod.Spec.Containers[1].VolumeMounts[0].Name,
				)
				require.Equal(t, "vcs", pod.Spec.Containers[1].VolumeMounts[1].Name)
				require.Equal(
					t,
					"docker-socket",
					pod.Spec.Containers[1].VolumeMounts[2].Name,
				)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			substrate := testCase.setup()
			err := substrate.createJobPod(
				context.Background(),
				testProject,
				testEvent,
				testJobName,
				testJobSpec,
			)
			testCase.assertions(substrate.kubeClient, err)
		})
	}
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

type mockQueueWriterFactory struct {
	NewWriterFn func(queueName string) (queue.Writer, error)
	CloseFn     func(context.Context) error
}

func (m *mockQueueWriterFactory) NewWriter(
	queueName string,
) (queue.Writer, error) {
	return m.NewWriterFn(queueName)
}

func (m *mockQueueWriterFactory) Close(ctx context.Context) error {
	return m.CloseFn(ctx)
}

type mockQueueWriter struct {
	WriteFn func(context.Context, string) error
	CloseFn func(context.Context) error
}

func (m *mockQueueWriter) Write(ctx context.Context, msg string) error {
	return m.WriteFn(ctx, msg)
}

func (m *mockQueueWriter) Close(ctx context.Context) error {
	return m.CloseFn(ctx)
}
