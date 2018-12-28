package vacuum

import (
	"log"
	"sort"
	"time"

	"github.com/Azure/brigade/pkg/storage"
	"github.com/Azure/brigade/pkg/storage/kube"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// NoMaxBuilds indicates that there is no maximum number of builds.
const NoMaxBuilds = -1

// NoMaxAge indicates that there is no maximum age.
var NoMaxAge = time.Time{}

const (
	buildFilter = "component = build, heritage = brigade"
)

// Vacuum describes a vacuum for cleaning up expired builds and jobs.
type Vacuum struct {
	age               time.Time
	max               int
	skipRunningBuilds bool
	namespace         string
	client            kubernetes.Interface
}

// New creates a new *Vacuum.
func New(age time.Time, max int, skipRunningBuilds bool, client kubernetes.Interface, ns string) *Vacuum {
	return &Vacuum{
		age:               age,
		max:               max,
		skipRunningBuilds: skipRunningBuilds,
		client:            client,
		namespace:         ns,
	}
}

// Run executes the vacuum, destroying resources that are expired.
func (v *Vacuum) Run() error {
	opts := metav1.ListOptions{
		LabelSelector: buildFilter,
	}

	if !v.age.IsZero() {
		log.Printf("Pruning records older than %s", v.age)
		secrets, err := v.client.CoreV1().Secrets(v.namespace).List(opts)
		if err != nil {
			return err
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
			}
		}
	}

	// If no max, return now.
	if v.max == NoMaxBuilds {
		return nil
	}

	// We need to re-load the secrets list and see if we are still over the max.
	secrets, err := v.client.CoreV1().Secrets(v.namespace).List(opts)
	if err != nil {
		return err
	}
	l := len(secrets.Items)
	if l <= v.max {
		log.Printf("Skipping vacuum. %d is ≤ max %d", l, v.max)
		return nil
	}
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
	}

	return nil
}

func (v *Vacuum) deleteBuild(bid string) error {
	store := kube.New(v.client, v.namespace)
	return store.DeleteBuild(bid, storage.DeleteBuildOptions{
		SkipRunningBuilds: v.skipRunningBuilds,
	})
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
