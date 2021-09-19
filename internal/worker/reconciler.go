package worker

import (
	"context"
	"fmt"
	"sort"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	batchlisters "k8s.io/client-go/listers/batch/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	utils "github.com/supercaracal/aws-sqs-worker-job-controller/internal/util"
	customapiv1 "github.com/supercaracal/aws-sqs-worker-job-controller/pkg/apis/supercaracal/v1"
	clientset "github.com/supercaracal/aws-sqs-worker-job-controller/pkg/generated/clientset/versioned"
	listers "github.com/supercaracal/aws-sqs-worker-job-controller/pkg/generated/listers/supercaracal/v1"
)

const (
	defaultHistoryLimit int32 = 10
)

// Reconciler is
type Reconciler struct {
	kubeClientSet        kubernetes.Interface
	customClientSet      clientset.Interface
	jobLister            batchlisters.JobLister
	customResourceLister listers.AWSSQSWorkerJobLister
	workQueue            workqueue.RateLimitingInterface
	recorder             record.EventRecorder
}

// NewReconciler is
func NewReconciler(
	kubeClientSet kubernetes.Interface,
	customClientSet clientset.Interface,
	jobLister batchlisters.JobLister,
	customResourceLister listers.AWSSQSWorkerJobLister,
	workQueue workqueue.RateLimitingInterface,
	recorder record.EventRecorder,
) *Reconciler {

	return &Reconciler{
		kubeClientSet:        kubeClientSet,
		customClientSet:      customClientSet,
		jobLister:            jobLister,
		customResourceLister: customResourceLister,
		workQueue:            workQueue,
		recorder:             recorder,
	}
}

// Run is
func (r *Reconciler) Run() {
	for r.processNextWorkItem() {
	}
}

func (r *Reconciler) processNextWorkItem() bool {
	obj, shutdown := r.workQueue.Get()
	if shutdown {
		return false
	}

	err := r.tryCleanupFinishedChildren(obj)
	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

func (r *Reconciler) tryCleanupFinishedChildren(obj interface{}) error {
	defer r.workQueue.Done(obj)

	var key string
	var ok bool

	if key, ok = obj.(string); !ok {
		r.workQueue.Forget(obj)
		utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
		return nil
	}

	if err := r.cleanupFinishedChildren(key); err != nil {
		r.workQueue.AddRateLimited(key)
		return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
	}

	r.workQueue.Forget(obj)

	return nil
}

func (r *Reconciler) cleanupFinishedChildren(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	parent, err := r.customResourceLister.AWSSQSWorkerJobs(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			utilruntime.HandleError(
				fmt.Errorf("custom resource '%s' in work queue no longer exists", key),
			)
			return nil
		}

		return err
	}

	children, err := r.jobLister.Jobs(namespace).List(labels.Everything())
	if errors.IsNotFound(err) {
		utilruntime.HandleError(fmt.Errorf("there is no jobs for custom resource '%s'", key))
		return nil
	}
	klog.V(4).Infof("Found %d jobs", len(children))

	ownedChildren := utils.ExtractOwnedChildren(children, parent)
	if len(ownedChildren) == 0 {
		return nil
	}
	klog.V(4).Infof("Found %d owned jobs", len(ownedChildren))

	sort.Sort(utils.AscJobs(ownedChildren))
	historyLimit := defaultHistoryLimit
	if parent.Spec.HistoryLimit != nil {
		historyLimit = *parent.Spec.HistoryLimit
	}
	size := int32(len(ownedChildren))
	if size > historyLimit {
		r.deleteChildren(ownedChildren[0:size-historyLimit], parent)
	}
	r.updateCustomResourceStatus(parent, ownedChildren[size-1])

	return nil
}

func (r *Reconciler) deleteChildren(children []*batchv1.Job, parent *customapiv1.AWSSQSWorkerJob) {
	for _, child := range children {
		isFinished, _ := utils.GetFinishedStatus(child)
		if !isFinished {
			continue
		}

		background := metav1.DeletePropagationBackground
		err := r.kubeClientSet.BatchV1().Jobs(child.Namespace).Delete(
			context.TODO(),
			child.Name,
			metav1.DeleteOptions{PropagationPolicy: &background},
		)

		if err != nil {
			r.recorder.Eventf(parent, corev1.EventTypeWarning, "Failed Delete", "Tried to deleted job %v", err)
			klog.Errorf("Error deleting job %s from %s/%s: %v", child.Name, parent.Namespace, parent.Name, err)
		} else {
			r.recorder.Eventf(parent, corev1.EventTypeNormal, "Successful Delete", "Deleted job %v", child.Name)
		}
	}
}

func (r *Reconciler) updateCustomResourceStatus(
	parent *customapiv1.AWSSQSWorkerJob,
	child *batchv1.Job,
) error {

	if child == nil {
		return nil
	}

	cpy := parent.DeepCopy()
	cpy.Status.StartTime = child.Status.StartTime
	cpy.Status.CompletionTime = child.Status.CompletionTime
	if _, finishedStatus := utils.GetFinishedStatus(child); finishedStatus == batchv1.JobComplete {
		cpy.Status.Succeeded = true
	} else {
		cpy.Status.Succeeded = false
	}

	_, err := r.customClientSet.AwssqsworkerjobcontrollerV1().
		AWSSQSWorkerJobs(parent.Namespace).
		Update(context.TODO(), cpy, metav1.UpdateOptions{})

	return err
}
