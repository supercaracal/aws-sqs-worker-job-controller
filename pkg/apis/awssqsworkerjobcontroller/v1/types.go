package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient

// AwsSqsWorkerJob is
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
type AwsSqsWorkerJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AwsSqsWorkerJobSpec   `json:"spec"`
	Status AwsSqsWorkerJobStatus `json:"status"`
}

// AwsSqsWorkerJobSpec is
type AwsSqsWorkerJobSpec struct {
	QueueURL string                 `json:"queueUrl"`
	Template corev1.PodTemplateSpec `json:"template"`
}

// AwsSqsWorkerJobStatus is
type AwsSqsWorkerJobStatus struct {
	StartTime      *metav1.Time
	CompletionTime *metav1.Time
	Succeeded      bool
}

// AwsSqsWorkerJobList is
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type AwsSqsWorkerJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []AwsSqsWorkerJob `json:"items"`
}
