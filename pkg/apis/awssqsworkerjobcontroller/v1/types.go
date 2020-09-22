package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	api "k8s.io/kubernetes/pkg/apis/core"
)

// +genclient

// AwsSqsWorkerJob is
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type AwsSqsWorkerJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AwsSqsWorkerJobSpec   `json:"spec"`
	Status AwsSqsWorkerJobStatus `json:"status"`
}

// AwsSqsWorkerJobSpec is
type AwsSqsWorkerJobSpec struct {
	QueueURL string              `json:"queueUrl"`
	Template api.PodTemplateSpec `json:"template"`
}

// AwsSqsWorkerJobStatus is
type AwsSqsWorkerJobStatus struct {
	Queues map[string]struct{} `json:"queues"`
}

// AwsSqsWorkerJobList is
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type AwsSqsWorkerJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []AwsSqsWorkerJob `json:"items"`
}
