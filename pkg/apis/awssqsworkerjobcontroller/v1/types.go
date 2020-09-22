package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	batch "k8s.io/kubernetes/pkg/apis/batch"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AwsSqsWorkerJob is
type AwsSqsWorkerJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              AwsSqsWorkerJobSpec   `json:"spec"`
	Status            AwsSqsWorkerJobStatus `json:"status"`
}

// @see https://github.com/kubernetes/gengo/blob/7794989d00002eae09b50e95c3a221245260a20e/examples/deepcopy-gen/generators/deepcopy.go#L843-L886
// @see https://github.com/kubernetes/apimachinery/blob/714f1137f89bf0ec6d038cf852d7661a1b9c660a/pkg/runtime/testing/types.go#L127-L156
// @see https://godoc.org/k8s.io/gengo/examples/deepcopy-gen

// AwsSqsWorkerJobSpec is
type AwsSqsWorkerJobSpec struct {
	JobTemplate JobTemplateSpec `json:"jobTemplate"`
	QueueName   string          `json:"queueName"`
}

// JobTemplateSpec is
// +k8s:deepcopy-gen=false
type JobTemplateSpec struct {
	batch.JobTemplateSpec `json:",inline"`
}

// AwsSqsWorkerJobStatus is
type AwsSqsWorkerJobStatus struct {
	QueueNames []string `json:"queueNames"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AwsSqsWorkerJobList is
type AwsSqsWorkerJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []AwsSqsWorkerJob `json:"items"`
}
