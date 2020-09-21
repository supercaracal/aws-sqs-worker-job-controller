package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	jobconfig "k8s.io/kubernetes/pkg/controller/job/config"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WorkerJob is
type WorkerJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WorkerJobSpec   `json:"spec"`
	Status WorkerJobStatus `json:"status"`

	JobSpec jobconfig.JobControllerConfiguration
}

// WorkerJobSpec is
type WorkerJobSpec struct {
	Queue string `json:"queue"`
}

// WorkerJobStatus is
type WorkerJobStatus struct {
	AvailableWorkers int32 `json:"availableWorkers"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WorkerJobList is
type WorkerJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []WorkerJob `json:"items"`
}
