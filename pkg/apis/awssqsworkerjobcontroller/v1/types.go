package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WorkerJob is a specification for a WorkerJob resource
type WorkerJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WorkerJobSpec   `json:"spec"`
	Status WorkerJobStatus `json:"status"`
}

// WorkerJobSpec is the spec for a WorkerJob resource
type WorkerJobSpec struct {
	DeploymentName string `json:"deploymentName"`
	Replicas       *int32 `json:"replicas"`
}

// WorkerJobStatus is the status for a WorkerJob resource
type WorkerJobStatus struct {
	AvailableReplicas int32 `json:"availableReplicas"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WorkerJobList is a list of WorkerJob resources
type WorkerJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []WorkerJob `json:"items"`
}
