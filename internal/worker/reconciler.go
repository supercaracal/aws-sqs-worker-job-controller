package worker

import (
	"k8s.io/client-go/kubernetes"
	batchlisterv1 "k8s.io/client-go/listers/batch/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	queues "github.com/supercaracal/aws-sqs-worker-job-controller/internal/queue"
	customclient "github.com/supercaracal/aws-sqs-worker-job-controller/pkg/generated/clientset/versioned"
	customlisterv1 "github.com/supercaracal/aws-sqs-worker-job-controller/pkg/generated/listers/supercaracal/v1"
)

// Reconciler is
type Reconciler struct {
	client       *ResourceClient
	lister       *ResourceLister
	workQueue    workqueue.RateLimitingInterface
	recorder     record.EventRecorder
	messageQueue queues.MessageQueue
}

// ResourceClient is
type ResourceClient struct {
	Builtin kubernetes.Interface
	Custom  customclient.Interface
}

// ResourceLister is
type ResourceLister struct {
	Job            batchlisterv1.JobLister
	CustomResource customlisterv1.AWSSQSWorkerJobLister
}

// NewReconciler is
func NewReconciler(
	cli *ResourceClient,
	list *ResourceLister,
	wq workqueue.RateLimitingInterface,
	rec record.EventRecorder,
) *Reconciler {

	return &Reconciler{client: cli, lister: list, workQueue: wq, recorder: rec}
}
