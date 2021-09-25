package worker

import (
	"context"
	"sort"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/klog/v2"

	customapiv1 "github.com/supercaracal/aws-sqs-worker-job-controller/pkg/apis/supercaracal/v1"
)

const (
	defaultHistoryLimit = 10
)

var (
	delOpts = metav1.DeleteOptions{PropagationPolicy: func(s metav1.DeletionPropagation) *metav1.DeletionPropagation { return &s }(metav1.DeletePropagationBackground)}
	updOpts = metav1.UpdateOptions{}
)

// Clean is
func (r *Reconciler) Clean() {
	parents, err := r.lister.CustomResource.List(labels.Everything())
	if err != nil {
		if !kubeerrors.IsNotFound(err) {
			utilruntime.HandleError(err)
		}
		return
	}

	allJobs, err := r.lister.Job.List(labels.Everything())
	if err != nil {
		if !kubeerrors.IsNotFound(err) {
			utilruntime.HandleError(err)
		}
		return
	}

	sort.Sort(JobsOrderedByStartTimeASC(allJobs))

	var historyLimit int
	for _, parent := range parents {
		historyLimit = defaultHistoryLimit
		if parent.Spec.HistoryLimit != nil {
			historyLimit = int(*parent.Spec.HistoryLimit)
		}

		children := extractChildren(parent, allJobs, historyLimit+4)
		size := len(children)

		if size == 0 {
			continue
		}

		if err := r.updateParent(parent, children[size-1]); err != nil {
			utilruntime.HandleError(err)
			continue
		}

		if size <= historyLimit {
			continue
		}

		for _, child := range children[0 : size-historyLimit] {
			if err := r.client.Builtin.BatchV1().Jobs(parent.Namespace).Delete(context.TODO(), child.Name, delOpts); err != nil {
				utilruntime.HandleError(err)
				r.recorder.Eventf(parent, corev1.EventTypeWarning, "Failure Delete", "Tried to deleted job %v", err)
				continue
			}

			r.recorder.Eventf(parent, corev1.EventTypeNormal, "Successful Delete", "Deleted job %v", child.Name)
			klog.V(4).Infof("Deleted resource %s/%s successfully", child.Namespace, child.Name)
		}
	}
}

func (r *Reconciler) updateParent(parent *customapiv1.AWSSQSWorkerJob, child *batchv1.Job) (err error) {
	if child == nil {
		return nil
	}

	cpy := parent.DeepCopy()
	cpy.Status.StartTime = child.Status.StartTime
	cpy.Status.CompletionTime = child.Status.CompletionTime
	cpy.Status.Succeeded = false
	if finishedStatus := getJobFinishedStatus(child); finishedStatus == batchv1.JobComplete {
		cpy.Status.Succeeded = true
	}

	_, err = r.client.Custom.SupercaracalV1().AWSSQSWorkerJobs(parent.Namespace).Update(context.TODO(), cpy, updOpts)
	return
}

func extractChildren(parent *customapiv1.AWSSQSWorkerJob, jobs []*batchv1.Job, size int) []*batchv1.Job {
	children := make([]*batchv1.Job, 0, size)
	for _, job := range jobs {
		if getJobFinishedStatus(job) == "" || !metav1.IsControlledBy(job, parent) {
			continue
		}

		children = append(children, job)
	}

	return children
}

func getJobFinishedStatus(job *batchv1.Job) batchv1.JobConditionType {
	for _, c := range job.Status.Conditions {
		if (c.Type == batchv1.JobComplete || c.Type == batchv1.JobFailed) && c.Status == corev1.ConditionTrue {
			return c.Type
		}
	}
	return ""
}
