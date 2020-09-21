package workers

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	customapiv1 "github.com/supercaracal/aws-sqs-worker-job-controller/pkg/apis/awssqsworkerjobcontroller/v1"
	clientset "github.com/supercaracal/aws-sqs-worker-job-controller/pkg/generated/clientset/versioned"
	listers "github.com/supercaracal/aws-sqs-worker-job-controller/pkg/generated/listers/awssqsworkerjobcontroller/v1"
)

const (
	// SuccessSynced is
	SuccessSynced = "Synced"
	// ErrResourceExists is
	ErrResourceExists = "ErrResourceExists"
	// MessageResourceExists is
	MessageResourceExists = "Resource %q already exists and is not managed by WorkerJob"
	// MessageResourceSynced is
	MessageResourceSynced = "WorkerJob synced successfully"
)

// ControllerWorker is
type ControllerWorker struct {
	kubeClientSet   kubernetes.Interface
	customClientSet clientset.Interface
	workerJobLister listers.WorkerJobLister
	workQueue       workqueue.RateLimitingInterface
	recorder        record.EventRecorder
}

// NewControllerWorker is
func NewControllerWorker(
	kubeClientSet kubernetes.Interface,
	customClientSet clientset.Interface,
	workerJobLister listers.WorkerJobLister,
	workQueue workqueue.RateLimitingInterface,
	recorder record.EventRecorder,
) *ControllerWorker {
	return &ControllerWorker{
		kubeClientSet:   kubeClientSet,
		customClientSet: customClientSet,
		workerJobLister: workerJobLister,
		workQueue:       workQueue,
		recorder:        recorder,
	}
}

// RunWorker is
func (w *ControllerWorker) RunWorker() {
	for w.processNextWorkItem() {
	}
}

func (w *ControllerWorker) processNextWorkItem() bool {
	obj, shutdown := w.workQueue.Get()
	if shutdown {
		return false
	}

	err := w.trySyncHandler(obj)
	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

func (w *ControllerWorker) trySyncHandler(obj interface{}) error {
	defer w.workQueue.Done(obj)

	var key string // format: namespace/name
	var ok bool
	if key, ok = obj.(string); !ok {
		w.workQueue.Forget(obj)
		utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
		return nil
	}

	if err := w.syncHandler(key); err != nil {
		w.workQueue.AddRateLimited(key)
		return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
	}

	w.workQueue.Forget(obj)
	klog.Infof("Successfully synced '%s'", key)

	return nil
}

func (w *ControllerWorker) syncHandler(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	workerJob, err := w.workerJobLister.WorkerJobs(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("workerjob '%s' in work queue no longer exists", key))
			return nil
		}

		return err
	}

	err = w.updateWorkJobStatus(workerJob)
	if err != nil {
		return err
	}

	w.recorder.Event(workerJob, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}

func (w *ControllerWorker) updateWorkJobStatus(
	workerJob *customapiv1.WorkerJob,
	deployment *appsv1.Deployment,
) error {
	workerJobCopy := workerJob.DeepCopy()
	workerJobCopy.Status.RegisteredWorkers++
	// If the CustomResourceSubresources feature gate is not enabled,
	// we must use Update instead of UpdateStatus to update the Status block of the Foo resource.
	// UpdateStatus will not allow changes to the Spec of the resource,
	// which is ideal for ensuring nothing other than resource status has been updated.
	_, err := w.customClientSet.AwssqsworkerjobcontrollerV1().WorkerJobs(workerJob.Namespace).Update(context.TODO(), workerJobCopy, metav1.UpdateOptions{})
	return err
}
