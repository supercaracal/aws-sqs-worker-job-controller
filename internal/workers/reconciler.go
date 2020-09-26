package workers

import (
	"context"
	"fmt"

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

	customapiv1 "github.com/supercaracal/aws-sqs-worker-job-controller/pkg/apis/awssqsworkerjobcontroller/v1"
	clientset "github.com/supercaracal/aws-sqs-worker-job-controller/pkg/generated/clientset/versioned"
	listers "github.com/supercaracal/aws-sqs-worker-job-controller/pkg/generated/listers/awssqsworkerjobcontroller/v1"
)

// Reconciler is
type Reconciler struct {
	kubeClientSet        kubernetes.Interface
	customClientSet      clientset.Interface
	jobLister            batchlisters.JobLister
	customResourceLister listers.AwsSqsWorkerJobLister
	workQueue            workqueue.RateLimitingInterface
	recorder             record.EventRecorder
}

// NewReconciler is
func NewReconciler(
	kubeClientSet kubernetes.Interface,
	customClientSet clientset.Interface,
	jobLister batchlisters.JobLister,
	customResourceLister listers.AwsSqsWorkerJobLister,
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

	err := r.tryCleanupFinishedJobs(obj)
	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

func (r *Reconciler) tryCleanupFinishedJobs(obj interface{}) error {
	defer r.workQueue.Done(obj)

	var key string
	var ok bool

	if key, ok = obj.(string); !ok {
		r.workQueue.Forget(obj)
		utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
		return nil
	}

	if err := r.cleanupFinishedJobs(key); err != nil {
		r.workQueue.AddRateLimited(key)
		return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
	}

	r.workQueue.Forget(obj)
	klog.Infof("Successfully cleaned children '%s'", key)

	return nil
}

func (r *Reconciler) cleanupFinishedJobs(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	parent, err := r.customResourceLister.AwsSqsWorkerJobs(namespace).Get(name)
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

	ownedChildren := extractOwnedChildren(children, parent)
	if len(ownedChildren) == 0 {
		return nil
	}
	klog.V(4).Infof("Found %d owned jobs", len(ownedChildren))

	lastChild := getLastFinishedChild(ownedChildren)
	r.updateCustomResourceStatus(parent, lastChild)

	if empty := r.deleteChildren(ownedChildren, parent); !empty {
		r.workQueue.Add(key)
		return nil
	}

	return nil
}

func (r *Reconciler) updateCustomResourceStatus(
	parent *customapiv1.AwsSqsWorkerJob,
	child *batchv1.Job,
) error {

	if child == nil {
		return nil
	}

	cpy := parent.DeepCopy()
	cpy.Status.StartTime = child.Status.StartTime
	cpy.Status.CompletionTime = child.Status.CompletionTime
	if _, finishedStatus := getFinishedStatus(child); finishedStatus == batchv1.JobComplete {
		cpy.Status.Succeeded = true
	} else {
		cpy.Status.Succeeded = false
	}

	_, err := r.customClientSet.AwssqsworkerjobcontrollerV1().
		AwsSqsWorkerJobs(parent.Namespace).
		Update(context.TODO(), cpy, metav1.UpdateOptions{})

	return err
}

func (r *Reconciler) deleteChildren(
	children []*batchv1.Job,
	parent *customapiv1.AwsSqsWorkerJob,
) bool {

	var empty bool = true

	for _, child := range children {
		isFinished, _ := getFinishedStatus(child)
		if !isFinished {
			empty = false
			continue
		}

		background := metav1.DeletePropagationBackground
		err := r.kubeClientSet.BatchV1().Jobs(child.Namespace).Delete(
			context.TODO(),
			child.Name,
			metav1.DeleteOptions{PropagationPolicy: &background},
		)
		if err != nil {
			r.recorder.Eventf(parent, corev1.EventTypeWarning, "FailedDelete", "Deleted job: %v", err)
			klog.Errorf(
				"Error deleting job %s from %s/%s: %v",
				child.Name,
				parent.Namespace,
				parent.Name,
				err,
			)
			empty = false
		}
		r.recorder.Eventf(parent, corev1.EventTypeNormal, "SuccessfulDelete", "Deleted job %v", child.Name)
	}

	return empty
}

func extractOwnedChildren(children []*batchv1.Job, parent *customapiv1.AwsSqsWorkerJob) []*batchv1.Job {
	ownedChildren := []*batchv1.Job{}

	for _, child := range children {
		if metav1.IsControlledBy(child, parent) {
			ownedChildren = append(ownedChildren, child)
		}
	}

	return ownedChildren
}

func getLastFinishedChild(children []*batchv1.Job) *batchv1.Job {
	var last *batchv1.Job = nil

	for _, child := range children {
		isFinished, _ := getFinishedStatus(child)
		if !isFinished {
			continue
		}
		if last == nil {
			last = child
			continue
		}
		if child.Status.StartTime.Before(last.Status.StartTime) {
			continue
		}
		last = child
	}

	return last
}

func getFinishedStatus(child *batchv1.Job) (bool, batchv1.JobConditionType) {
	for _, c := range child.Status.Conditions {
		if (c.Type == batchv1.JobComplete || c.Type == batchv1.JobFailed) && c.Status == corev1.ConditionTrue {
			return true, c.Type
		}
	}
	return false, ""
}
