package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/brigadecore/brigade/v2/apiserver/internal/core"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/queue"
	myk8s "github.com/brigadecore/brigade/v2/internal/kubernetes"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/kubernetes"
)

var workerPodsSelector = labels.Set(
	map[string]string{
		myk8s.LabelComponent: "worker",
	},
).AsSelector().String()

var jobPodsSelector = labels.Set(
	map[string]string{
		myk8s.LabelComponent: "job",
	},
).AsSelector().String()

var runningPodsSelector = fields.Set(
	map[string]string{
		"status.phase": string(corev1.PodRunning),
	},
).AsSelector().String()

var pendingPodsSelector = fields.Set(
	map[string]string{
		"status.phase": string(corev1.PodPending),
	},
).AsSelector().String()

var unknownPhasePodsSelector = fields.Set(
	map[string]string{
		"status.phase": string(corev1.PodUnknown),
	},
).AsSelector().String()

// SubstrateConfig encapsulates several configuration options for the
// Kubernetes-based Substrate.
type SubstrateConfig struct {
	// APIAddress is the address of the Brigade API server. The substrate will use
	// this information whenever it needs to tell another component where to find
	// the API server.
	APIAddress string
}

// substrate is a Kubernetes-based implementation of the core.Substrate
// interface.
type substrate struct {
	generateNewNamespaceFn func(projectID string) string
	kubeClient             kubernetes.Interface
	queueWriterFactory     queue.WriterFactory
	config                 SubstrateConfig
}

// NewSubstrate returns a Kubernetes-based implementation of the core.Substrate
// interface.
func NewSubstrate(
	kubeClient kubernetes.Interface,
	queueWriterFactory queue.WriterFactory,
	config SubstrateConfig,
) core.Substrate {
	return &substrate{
		generateNewNamespaceFn: generateNewNamespace,
		kubeClient:             kubeClient,
		queueWriterFactory:     queueWriterFactory,
		config:                 config,
	}
}

func (s *substrate) CountRunningWorkers(
	ctx context.Context,
) (core.SubstrateWorkerCount, error) {
	count := core.SubstrateWorkerCount{}
	var err error
	count.Count, err = s.countRunningPods(ctx, workerPodsSelector)
	return count, err
}

func (s *substrate) CountRunningJobs(
	ctx context.Context,
) (core.SubstrateJobCount, error) {
	count := core.SubstrateJobCount{}
	var err error
	count.Count, err = s.countRunningPods(ctx, jobPodsSelector)
	return count, err
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

func (s *substrate) ScheduleWorker(
	ctx context.Context,
	project core.Project,
	event core.Event,
) error {
	// Create a Kubernetes secret containing relevant Event and Project details.
	// This is created PRIOR to scheduling so that these details will reflect an
	// accurate snapshot of Project configuration at the time the Event was
	// created.

	projectSecretsSecret, err := s.kubeClient.CoreV1().Secrets(
		project.Kubernetes.Namespace,
	).Get(ctx, "project-secrets", metav1.GetOptions{})
	if err != nil {
		return errors.Wrapf(
			err,
			"error finding secret \"project-secrets\" in namespace %q",
			project.Kubernetes.Namespace,
		)
	}
	secrets := map[string]string{}
	for key, value := range projectSecretsSecret.Data {
		secrets[key] = string(value)
	}

	type proj struct {
		ID         string                 `json:"id"`
		Kubernetes core.KubernetesDetails `json:"kubernetes"`
		Secrets    map[string]string      `json:"secrets"`
	}

	type worker struct {
		APIAddress           string            `json:"apiAddress"`
		APIToken             string            `json:"apiToken"`
		LogLevel             core.LogLevel     `json:"logLevel"`
		ConfigFilesDirectory string            `json:"configFilesDirectory"`
		DefaultConfigFiles   map[string]string `json:"defaultConfigFiles" bson:"defaultConfigFiles"` // nolint: lll
	}

	// Create a secret with event details
	eventJSON, err := json.MarshalIndent(
		struct {
			ID         string `json:"id"`
			Project    proj   `json:"project"`
			Source     string `json:"source"`
			Type       string `json:"type"`
			ShortTitle string `json:"shortTitle"`
			LongTitle  string `json:"longTitle"`
			Payload    string `json:"payload"`
			Worker     worker `json:"worker"`
		}{
			ID: event.ID,
			Project: proj{
				ID:         event.ProjectID,
				Kubernetes: *project.Kubernetes,
				Secrets:    secrets,
			},
			Source:     event.Source,
			Type:       event.Type,
			ShortTitle: event.ShortTitle,
			LongTitle:  event.LongTitle,
			Payload:    event.Payload,
			Worker: worker{
				APIAddress:           s.config.APIAddress,
				APIToken:             event.Worker.Token,
				LogLevel:             event.Worker.Spec.LogLevel,
				ConfigFilesDirectory: event.Worker.Spec.ConfigFilesDirectory,
				DefaultConfigFiles:   event.Worker.Spec.DefaultConfigFiles,
			},
		},
		"",
		"  ",
	)
	if err != nil {
		return errors.Wrapf(err, "error marshaling event %q", event.ID)
	}

	data := map[string][]byte{}
	data["event.json"] = eventJSON
	// Special treatment for secrets named gitSSHKey and gitSSHCert. If they're
	// defined, add them to this secret so the worker's VCS init container (if
	// applicable) has easy access to them.
	if gitSSHKey, ok := projectSecretsSecret.Data["gitSSHKey"]; ok {
		data["gitSSHKey"] = gitSSHKey
	}
	if gitSSHCert, ok := projectSecretsSecret.Data["gitSSHCert"]; ok {
		data["gitSSHCert"] = gitSSHCert
	}

	if _, err = s.kubeClient.CoreV1().Secrets(
		project.Kubernetes.Namespace,
	).Create(
		ctx,
		&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("event-%s", event.ID),
				Labels: map[string]string{
					myk8s.LabelComponent: "event",
					myk8s.LabelProject:   event.ProjectID,
					myk8s.LabelEvent:     event.ID,
				},
			},
			Type: myk8s.SecretTypeEvent,
			Data: data,
		},
		metav1.CreateOptions{},
	); err != nil {
		return errors.Wrapf(
			err,
			"error creating secret %q in namespace %q",
			event.ID,
			project.Kubernetes.Namespace,
		)
	}

	queueWriter, err := s.queueWriterFactory.NewWriter(
		fmt.Sprintf("workers.%s", event.ProjectID),
	)
	if err != nil {
		return errors.Wrapf(
			err,
			"error creating queue writer for project %q workers",
			event.ProjectID,
		)
	}
	defer func() {
		closeCtx, cancelCloseCtx :=
			context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelCloseCtx()
		queueWriter.Close(closeCtx)
	}()

	if err := queueWriter.Write(ctx, event.ID); err != nil {
		return errors.Wrapf(
			err,
			"error submitting execution task for event %q worker",
			event.ID,
		)
	}

	return nil
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

func generateNewNamespace(projectID string) string {
	return fmt.Sprintf("brigade-%s-%s", projectID, uuid.NewV4().String())
}

func (s *substrate) countRunningPods(
	ctx context.Context,
	labelSelector string,
) (int, error) {
	phaseSelectors := []string{
		pendingPodsSelector,
		runningPodsSelector,
		unknownPhasePodsSelector,
	}
	// Use a set to assist in counting. This helps prevent us from double-counting
	// a pod if its phase changes from pending to running while we're counting.
	podsSet := map[string]struct{}{}
	for _, phaseSelector := range phaseSelectors {
		var cont string
		for {
			pods, err := s.kubeClient.CoreV1().Pods("").List(
				ctx,
				metav1.ListOptions{
					LabelSelector: labelSelector,
					FieldSelector: phaseSelector,
					Continue:      cont,
				},
			)
			if err != nil {
				return 0, errors.Wrap(err, "error counting pods")
			}
			for _, pod := range pods.Items {
				podsSet[fmt.Sprintf("%s:%s", pod.Namespace, pod.Name)] = struct{}{}
			}
			cont = pods.Continue
			if cont == "" {
				break
			}
		}
	}
	return len(podsSet), nil
}
