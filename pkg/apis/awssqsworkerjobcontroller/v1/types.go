package v1

import (
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WorkerJob is
type WorkerJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WorkerJobSpec   `json:"spec"`
	Status WorkerJobStatus `json:"status"`
}

// WorkerJobSpec is
type WorkerJobSpec struct {
	QueueName                  string `json:"queueName"`
	SuccessfulJobsHistoryLimit *int32 `json:"successfulJobsHistoryLimit,omitempty"`
	FailedJobsHistoryLimit     *int32 `json:"failedJobsHistoryLimit,omitempty"`
	//JobTemplate                batchv1.JobTemplateSpec `json:"jobTemplate"`
}

// WorkerJobStatus is
type WorkerJobStatus struct {
	RegisteredWorkers int32 `json:"registeredWorkers"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WorkerJobList is
type WorkerJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []WorkerJob `json:"items"`
}
