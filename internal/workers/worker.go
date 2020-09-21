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
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the Foo resource with this namespace/name
	foo, err := w.workerJobLister.WorkerJobs(namespace).Get(name)
	if err != nil {
		// The Foo resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("foo '%s' in work queue no longer exists", key))
			return nil
		}

		return err
	}

	deploymentName := foo.Spec.DeploymentName
	if deploymentName == "" {
		// We choose to absorb the error here as the worker would requeue the
		// resource otherwise. Instead, the next time the resource is updated
		// the resource will be queued again.
		utilruntime.HandleError(fmt.Errorf("%s: deployment name must be specified", key))
		return nil
	}

	// Get the deployment with the name specified in Foo.spec
	deployment, err := c.deploymentsLister.Deployments(foo.Namespace).Get(deploymentName)
	// If the resource doesn't exist, we'll create it
	if errors.IsNotFound(err) {
		deployment, err = w.kubeClientSet.AppsV1().Deployments(foo.Namespace).Create(context.TODO(), newDeployment(foo), metav1.CreateOptions{})
	}

	// If an error occurs during Get/Create, we'll requeue the item so we can
	// attempt processing again later. This could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		return err
	}

	// If the Deployment is not controlled by this Foo resource, we should log
	// a warning to the event recorder and return error msg.
	if !metav1.IsControlledBy(deployment, foo) {
		msg := fmt.Sprintf(MessageResourceExists, deployment.Name)
		w.recorder.Event(foo, corev1.EventTypeWarning, ErrResourceExists, msg)
		return fmt.Errorf(msg)
	}

	// If this number of the replicas on the Foo resource is specified, and the
	// number does not equal the current desired replicas on the Deployment, we
	// should update the Deployment resource.
	if foo.Spec.Replicas != nil && *foo.Spec.Replicas != *deployment.Spec.Replicas {
		klog.V(4).Infof("Foo %s replicas: %d, deployment replicas: %d", name, *foo.Spec.Replicas, *deployment.Spec.Replicas)
		deployment, err = w.kubeClientSet.AppsV1().Deployments(foo.Namespace).Update(context.TODO(), newDeployment(foo), metav1.UpdateOptions{})
	}

	// If an error occurs during Update, we'll requeue the item so we can
	// attempt processing again later. This could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		return err
	}

	// Finally, we update the status block of the Foo resource to reflect the
	// current state of the world
	err = w.updateFooStatus(foo, deployment)
	if err != nil {
		return err
	}

	w.recorder.Event(foo, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}

func (w *ControllerWorker) updateFooStatus(foo *customapiv1.WorkerJob, deployment *appsv1.Deployment) error {
	// NEVER modify objects from the store. It's a read-only, local cache.
	// You can use DeepCopy() to make a deep copy of original object and modify this copy
	// Or create a copy manually for better performance
	fooCopy := foo.DeepCopy()
	fooCopy.Status.AvailableReplicas = deployment.Status.AvailableReplicas
	// If the CustomResourceSubresources feature gate is not enabled,
	// we must use Update instead of UpdateStatus to update the Status block of the Foo resource.
	// UpdateStatus will not allow changes to the Spec of the resource,
	// which is ideal for ensuring nothing other than resource status has been updated.
	_, err := w.customClientSet.AwssqsworkerjobcontrollerV1().WorkerJobs(foo.Namespace).Update(context.TODO(), fooCopy, metav1.UpdateOptions{})
	return err
}

// newDeployment creates a new Deployment for a Foo resource. It also sets
// the appropriate OwnerReferences on the resource so handleObject can discover
// the Foo resource that 'owns' it.
func newDeployment(foo *customapiv1.WorkerJob) *appsv1.Deployment {
	labels := map[string]string{
		"app":        "nginx",
		"controller": foo.Name,
	}
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      foo.Spec.DeploymentName,
			Namespace: foo.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(foo, customapiv1.SchemeGroupVersion.WithKind("Foo")),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: foo.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "nginx",
							Image: "nginx:latest",
						},
					},
				},
			},
		},
	}
}
