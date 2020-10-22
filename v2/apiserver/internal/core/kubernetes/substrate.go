package kubernetes

import (
	"context"
	"fmt"

	"github.com/brigadecore/brigade/v2/apiserver/internal/core"
	myk8s "github.com/brigadecore/brigade/v2/internal/kubernetes"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/kubernetes"
)

// substrate is a Kubernetes-based implementation of the core.Substrate
// interface.
type substrate struct {
	generateNewNamespaceFn func(projectID string) string
	kubeClient             kubernetes.Interface
}

// NewSubstrate returns a Kubernetes-based implementation of the core.Substrate
// interface.
func NewSubstrate(
	kubeClient kubernetes.Interface,
) core.Substrate {
	return &substrate{
		generateNewNamespaceFn: generateNewNamespace,
		kubeClient:             kubeClient,
	}
}

func (s *substrate) CreateProject(
	ctx context.Context,
	project core.Project,
) (core.Project, error) {
	// Generate and assign a unique Kubernetes namespace name for the Project,
	// but don't create it yet
	project.Kubernetes = &core.KubernetesDetails{
		Namespace: s.generateNewNamespaceFn(project.ID),
	}

	// Create the Project's Kubernetes namespace
	if _, err := s.kubeClient.CoreV1().Namespaces().Create(
		ctx,
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: project.Kubernetes.Namespace,
			},
		},
		metav1.CreateOptions{},
	); err != nil {
		return project, errors.Wrapf(
			err,
			"error creating namespace %q for project %q",
			project.Kubernetes.Namespace,
			project.ID,
		)
	}

	// Create an RBAC Role for use by all the Project's Workers
	if _, err := s.kubeClient.RbacV1().Roles(project.Kubernetes.Namespace).Create(
		ctx,
		&rbacv1.Role{
			ObjectMeta: metav1.ObjectMeta{
				Name: "workers",
			},
			Rules: []rbacv1.PolicyRule{},
		},
		metav1.CreateOptions{},
	); err != nil {
		return project, errors.Wrapf(
			err,
			"error creating role \"workers\" in namespace %q",
			project.Kubernetes.Namespace,
		)
	}

	// Create a ServiceAccount for use by all of the Project's Workers
	if _, err := s.kubeClient.CoreV1().ServiceAccounts(
		project.Kubernetes.Namespace,
	).Create(
		ctx,
		&corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name: "workers",
			},
		},
		metav1.CreateOptions{},
	); err != nil {
		return project, errors.Wrapf(
			err,
			"error creating service account \"workers\" in namespace %q",
			project.Kubernetes.Namespace,
		)
	}

	// Create an RBAC RoleBinding to associate the Workers' ServiceAccount with
	// the Workers' RBAC Role
	if _, err := s.kubeClient.RbacV1().RoleBindings(
		project.Kubernetes.Namespace,
	).Create(
		ctx,
		&rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name: "workers",
			},
			Subjects: []rbacv1.Subject{
				{
					Kind:      "ServiceAccount",
					Name:      "workers",
					Namespace: project.Kubernetes.Namespace,
				},
			},
			RoleRef: rbacv1.RoleRef{
				Kind: "Role",
				Name: "workers",
			},
		},
		metav1.CreateOptions{},
	); err != nil {
		return project, errors.Wrapf(
			err,
			"error creating role binding \"workers\" in namespace %q",
			project.Kubernetes.Namespace,
		)
	}

	// Create an RBAC role for use by all of the Project's Jobs
	if _, err := s.kubeClient.RbacV1().Roles(project.Kubernetes.Namespace).Create(
		ctx,
		&rbacv1.Role{
			ObjectMeta: metav1.ObjectMeta{
				Name: "jobs",
			},
			Rules: []rbacv1.PolicyRule{},
		},
		metav1.CreateOptions{},
	); err != nil {
		return project, errors.Wrapf(
			err,
			"error creating role \"jobs\" in namespace %q",
			project.Kubernetes.Namespace,
		)
	}

	// Create a service account for use by all of the Project's Jobs
	if _, err := s.kubeClient.CoreV1().ServiceAccounts(
		project.Kubernetes.Namespace,
	).Create(
		ctx,
		&corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name: "jobs",
			},
		},
		metav1.CreateOptions{},
	); err != nil {
		return project, errors.Wrapf(
			err,
			"error creating service account \"jobs\" in namespace %q",
			project.Kubernetes.Namespace,
		)
	}

	// Create an RBAC role binding to associate the jobs service account with the
	// jobs RBAC role
	if _, err := s.kubeClient.RbacV1().RoleBindings(
		project.Kubernetes.Namespace,
	).Create(
		ctx,
		&rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name: "jobs",
			},
			Subjects: []rbacv1.Subject{
				{
					Kind:      "ServiceAccount",
					Name:      "jobs",
					Namespace: project.Kubernetes.Namespace,
				},
			},
			RoleRef: rbacv1.RoleRef{
				Kind: "Role",
				Name: "jobs",
			},
		},
		metav1.CreateOptions{},
	); err != nil {
		return project, errors.Wrapf(
			err,
			"error creating role binding \"jobs\" in namespace %q",
			project.Kubernetes.Namespace,
		)
	}

	// Create a Kubernetes Secret to store the Project's Secrets. Note that the
	// Kubernetes-based implementation of the SecretStore interface will assume
	// this Kubernetes secret exists.
	if _, err := s.kubeClient.CoreV1().Secrets(
		project.Kubernetes.Namespace,
	).Create(
		ctx,
		&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: "project-secrets",
				Labels: map[string]string{
					myk8s.LabelComponent: "project-secrets",
					myk8s.LabelProject:   project.ID,
				},
			},
			Type: myk8s.SecretTypeProjectSecrets,
		},
		metav1.CreateOptions{},
	); err != nil {
		return project, errors.Wrapf(
			err,
			"error creating secret \"project-secrets\" in namespace %q",
			project.Kubernetes.Namespace,
		)
	}

	return project, nil
}

func (s *substrate) DeleteProject(
	ctx context.Context,
	project core.Project,
) error {
	// Just delete the Project's entire Kubernetes namespace and it should take
	// all other Project resources along with it.
	if err := s.kubeClient.CoreV1().Namespaces().Delete(
		ctx,
		project.Kubernetes.Namespace,
		metav1.DeleteOptions{},
	); err != nil {
		return errors.Wrapf(
			err,
			"error deleting namespace %q",
			project.Kubernetes.Namespace,
		)
	}
	return nil
}

func generateNewNamespace(projectID string) string {
	return fmt.Sprintf("brigade-%s-%s", projectID, uuid.NewV4().String())
}

func (s *substrate) DeleteWorkerAndJobs(
	ctx context.Context,
	project core.Project,
	event core.Event,
) error {
	matchesEvent, _ := labels.NewRequirement(
		myk8s.LabelEvent,
		selection.Equals,
		[]string{event.ID},
	)
	labelSelector := labels.NewSelector()
	labelSelector = labelSelector.Add(*matchesEvent)

	// Delete all pods related to this Event
	if err := s.kubeClient.CoreV1().Pods(
		project.Kubernetes.Namespace,
	).DeleteCollection(
		ctx,
		metav1.DeleteOptions{},
		metav1.ListOptions{
			LabelSelector: labelSelector.String(),
		},
	); err != nil {
		return errors.Wrapf(
			err,
			"error deleting event %q pods in namespace %q",
			event.ID,
			project.Kubernetes.Namespace,
		)
	}

	// Delete all persistent volume claims related to this Event
	if err := s.kubeClient.CoreV1().PersistentVolumeClaims(
		project.Kubernetes.Namespace,
	).DeleteCollection(
		ctx,
		metav1.DeleteOptions{},
		metav1.ListOptions{
			LabelSelector: labelSelector.String(),
		},
	); err != nil {
		return errors.Wrapf(
			err,
			"error deleting event %q persistent volume claims in namespace %q",
			event.ID,
			project.Kubernetes.Namespace,
		)
	}

	// Delete all secrets related to this Event
	if err := s.kubeClient.CoreV1().Secrets(
		project.Kubernetes.Namespace,
	).DeleteCollection(
		ctx,
		metav1.DeleteOptions{},
		metav1.ListOptions{
			LabelSelector: labelSelector.String(),
		},
	); err != nil {
		return errors.Wrapf(
			err,
			"error deleting event %q secrets in namespace %q",
			event.ID,
			project.Kubernetes.Namespace,
		)
	}

	return nil
}
