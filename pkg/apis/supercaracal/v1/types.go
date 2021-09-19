package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient

// AWSSQSWorkerJob is
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type AWSSQSWorkerJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AWSSQSWorkerJobSpec   `json:"spec"`
	Status AWSSQSWorkerJobStatus `json:"status"`
}

// AWSSQSWorkerJobSpec is
type AWSSQSWorkerJobSpec struct {
	// The URL of the queue which is treated by the controller for tasks.
	QueueURL string `json:"queueURL"`

	// The number of finished jobs to retain.
	// +optional
	HistoryLimit *int32 `json:"historyLimit,omitempty"`

	// Defines pods that will be created from this template.
	Template corev1.PodTemplateSpec `json:"template"`
}

// AWSSQSWorkerJobStatus is
type AWSSQSWorkerJobStatus struct {
	StartTime      *metav1.Time
	CompletionTime *metav1.Time
	Succeeded      bool
}

// AWSSQSWorkerJobList is
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type AWSSQSWorkerJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []AWSSQSWorkerJob `json:"items"`
}
