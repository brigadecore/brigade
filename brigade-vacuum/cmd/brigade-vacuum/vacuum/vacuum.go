package vacuum

import (
	"fmt"
	"log"
	"sort"
	"time"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	buildFilter = "component = build, heritage = brigade"
	jobFilter   = "component in (build, job), heritage = brigade, build = %s"
)

// Vacuum describes a vacuum for cleaning up expired builds and jobs.
type Vacuum struct {
	age       time.Time
	max       int
	namespace string
	client    kubernetes.Interface
}

// New creates a new *Vacuum.
func New(age time.Time, max int, client kubernetes.Interface, ns string) *Vacuum {
	return &Vacuum{
		age:       age,
		max:       max,
		client:    client,
		namespace: ns,
	}
}

// Run executes the vacuum, destroying resources that are expired.
//
// It returns the number of builds deleted.
func (v *Vacuum) Run() (int, error) {
	opts := metav1.ListOptions{
		LabelSelector: buildFilter,
	}

	deleted := 0

	if !v.age.IsZero() {
		log.Printf("Pruning records older than %s", v.age)
		secrets, err := v.client.CoreV1().Secrets(v.namespace).List(opts)
		if err != nil {
			return 0, err
		}
		for _, s := range secrets.Items {
			ts := s.ObjectMeta.CreationTimestamp.Time
			bid, ok := s.ObjectMeta.Labels["build"]
			if !ok {
				log.Printf("Build %q has no build ID. Skipping.\n", s.Name)
				continue
			}
			if v.age.After(ts) {
				if err := v.deleteBuild(bid); err != nil {
					log.Printf("Failed to delete build %s: %s (age)\n", bid, err)
					continue
				}
				deleted++
			}
		}
	}

	// If no max, return now.
	if v.max == 0 {
		return deleted, nil
	}

	// We need to re-load the secrets list and see if we are still over the max.
	secrets, err := v.client.CoreV1().Secrets(v.namespace).List(opts)
	if err != nil {
		return deleted, err
	}
	l := len(secrets.Items)
	if l > v.max {
		sort.Sort(ByCreation(secrets.Items))
		for i := v.max; i < l; i++ {
			// Delete secret and builds
			s := secrets.Items[i]
			bid, ok := s.ObjectMeta.Labels["build"]
			if !ok {
				log.Printf("Build %q has no build ID. Skipping.\n", s.Name)
				continue
			}
			if err := v.deleteBuild(bid); err != nil {
				log.Printf("Failed to delete build %s: %s (max)\n", bid, err)
				continue
			}
			deleted++
		}
	}

	return deleted, nil
}

func (v *Vacuum) deleteBuild(bid string) error {
	opts := metav1.ListOptions{
		LabelSelector: fmt.Sprintf(jobFilter, bid),
	}
	delOpts := metav1.NewDeleteOptions(0)
	secrets, err := v.client.CoreV1().Secrets(v.namespace).List(opts)
	if err != nil {
		return err
	}
	for _, s := range secrets.Items {
		log.Printf("Deleting secret %q", s.Name)
		if err := v.client.CoreV1().Secrets(v.namespace).Delete(s.Name, delOpts); err != nil {
			log.Printf("failed to delete job secret %s (continuing): %s", s.Name, err)
		}
	}

	pods, err := v.client.CoreV1().Pods(v.namespace).List(opts)
	if err != nil {
		return err
	}
	for _, p := range pods.Items {
		log.Printf("Deleting pod %q", p.Name)
		if err := v.client.CoreV1().Pods(v.namespace).Delete(p.Name, delOpts); err != nil {
			log.Printf("failed to delete job pod %s (continuing): %s", p.Name, err)
		}
	}

	// As a safety condition, we might also consider deleting PVCs.
	return nil
}

// ByCreation sorts secrets by their creation timestamp.
type ByCreation []v1.Secret

// Len returns the length of the secrets slice.
func (b ByCreation) Len() int {
	return len(b)
}

// Swap swaps the position of two indices.
func (b ByCreation) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

// Less tests that i is less than j.
func (b ByCreation) Less(i, j int) bool {
	jj := b[j].ObjectMeta.CreationTimestamp.Time
	ii := b[i].ObjectMeta.CreationTimestamp.Time
	return ii.After(jj)
}
