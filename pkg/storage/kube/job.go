package kube

import (
	"bytes"
	"fmt"
	"io"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/pkg/api/v1"

	"github.com/deis/acid/pkg/acid"
)

func (s *store) GetJob(id string) (*acid.Job, error) {
	labels := labels.Set{"heritage": "acid"}
	listOption := meta.ListOptions{LabelSelector: labels.AsSelector().String()}
	pods, err := s.client.CoreV1().Pods(s.namespace).List(listOption)
	if err != nil {
		return nil, err
	}
	if len(pods.Items) < 1 {
		return nil, fmt.Errorf("could not find job %s: no pod exists with label %s", id, labels.AsSelector().String())
	}
	for i := range pods.Items {
		job := NewJobFromPod(pods.Items[i])
		if job.ID == id {
			return job, nil
		}
	}
	return nil, fmt.Errorf("could not find job %s: no pod exists with that ID and label %s", id, labels.AsSelector().String())
}

func (s *store) GetBuildJobs(build *acid.Build) ([]*acid.Job, error) {
	// Load the pods that ran as part of this build.
	lo := meta.ListOptions{LabelSelector: fmt.Sprintf("heritage=acid,component=job,commit=%s,project=%s", build.Commit, build.ProjectID)}

	podList, err := s.client.CoreV1().Pods(s.namespace).List(lo)
	if err != nil {
		return nil, err
	}
	jobList := make([]*acid.Job, len(podList.Items))
	for i := range podList.Items {
		jobList[i] = NewJobFromPod(podList.Items[i])
	}
	return jobList, nil
}
func (s *store) GetJobLogStream(job *acid.Job) (io.ReadCloser, error) {
	req := s.client.CoreV1().Pods(s.namespace).GetLogs(job.ID, &v1.PodLogOptions{})

	readCloser, err := req.Stream()
	if err != nil {
		return nil, err
	}
	return readCloser, nil
}
func (s *store) GetJobLog(job *acid.Job) (string, error) {
	buf := new(bytes.Buffer)
	r, err := s.GetJobLogStream(job)
	if err != nil {
		return "", err
	}
	defer r.Close()
	io.Copy(buf, r)
	return buf.String(), nil
}

// NewJobFromPod parses the pod object metadata and deserializes it into a Job.
func NewJobFromPod(pod v1.Pod) *acid.Job {
	job := &acid.Job{
		ID:           pod.ObjectMeta.Name,
		Name:         pod.ObjectMeta.Labels["jobname"],
		CreationTime: pod.ObjectMeta.CreationTimestamp.Time,
		Image:        pod.Spec.Containers[0].Image,
		Status:       acid.JobStatus(pod.Status.Phase),
	}

	if (job.Status != acid.JobPending) && (job.Status != acid.JobUnknown) {
		job.StartTime = pod.Status.StartTime.Time
	}

	if len(pod.Status.ContainerStatuses) > 0 {
		if pod.Status.ContainerStatuses[0].State.Terminated != nil {
			job.EndTime = pod.Status.ContainerStatuses[0].State.Terminated.FinishedAt.Time
			job.ExitCode = pod.Status.ContainerStatuses[0].State.Terminated.ExitCode
		}
	}

	return job
}
