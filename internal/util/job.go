package util

import (
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	customapiv1 "github.com/supercaracal/aws-sqs-worker-job-controller/pkg/apis/supercaracal/v1"
)

// AscJobs is
type AscJobs []*batchv1.Job

func (aj AscJobs) Len() int {
	return len(aj)
}

func (aj AscJobs) Swap(i, j int) {
	aj[i], aj[j] = aj[j], aj[i]
}

func (aj AscJobs) Less(i, j int) bool {
	if aj[i].Status.StartTime == nil && aj[j].Status.StartTime != nil {
		return false
	}

	if aj[i].Status.StartTime != nil && aj[j].Status.StartTime == nil {
		return true
	}

	if aj[i].Status.StartTime.Equal(aj[j].Status.StartTime) {
		return aj[i].Name < aj[j].Name
	}

	return aj[i].Status.StartTime.Before(aj[j].Status.StartTime)
}

// ExtractOwnedChildren is
func ExtractOwnedChildren(children []*batchv1.Job, parent *customapiv1.AWSSQSWorkerJob) []*batchv1.Job {
	ownedChildren := []*batchv1.Job{}

	for _, child := range children {
		if metav1.IsControlledBy(child, parent) {
			ownedChildren = append(ownedChildren, child)
		}
	}

	return ownedChildren
}

// GetFinishedStatus is
func GetFinishedStatus(child *batchv1.Job) (bool, batchv1.JobConditionType) {
	for _, c := range child.Status.Conditions {
		if (c.Type == batchv1.JobComplete || c.Type == batchv1.JobFailed) && c.Status == corev1.ConditionTrue {
			return true, c.Type
		}
	}
	return false, ""
}
