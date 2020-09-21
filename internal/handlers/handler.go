package handlers

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	listers "github.com/supercaracal/aws-sqs-worker-job-controller/pkg/generated/listers/awssqsworkerjobcontroller/v1"
)

// InformerHandler is
type InformerHandler struct {
	workerJobLister listers.WorkerJobLister
	workQueue       workqueue.RateLimitingInterface
}

// NewInformerHandler is
func NewInformerHandler(
	workerJobLister listers.WorkerJobLister,
	workQueue workqueue.RateLimitingInterface,
) *InformerHandler {
	return &InformerHandler{
		workerJobLister: workerJobLister,
		workQueue:       workQueue,
	}
}

// OnAdd is
func (h *InformerHandler) OnAdd(obj interface{}) {
	h.enqueueWorkerJob(obj)
}

// OnUpdate is
func (h *InformerHandler) OnUpdate(old, new interface{}) {
	h.enqueueWorkerJob(new)

	newDepl := new.(*appsv1.Deployment)
	oldDepl := old.(*appsv1.Deployment)
	if newDepl.ResourceVersion == oldDepl.ResourceVersion {
		return
	}
	h.handleObject(new)
}

// OnDelete is
func (h *InformerHandler) OnDelete(obj interface{}) {
	h.handleObject(obj)
}

func (h *InformerHandler) enqueueWorkerJob(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	h.workQueue.Add(key)
}

func (h *InformerHandler) handleObject(obj interface{}) {
	var object metav1.Object
	var ok bool
	if object, ok = obj.(metav1.Object); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("error decoding object, invalid type"))
			return
		}
		object, ok = tombstone.Obj.(metav1.Object)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("error decoding object tombstone, invalid type"))
			return
		}
		klog.V(4).Infof("Recovered deleted object '%s' from tombstone", object.GetName())
	}
	klog.V(4).Infof("Processing object: %s", object.GetName())
	if ownerRef := metav1.GetControllerOf(object); ownerRef != nil {
		// If this object is not owned by a Foo, we should not do anything more
		// with it.
		if ownerRef.Kind != "Foo" {
			return
		}

		foo, err := h.workerJobLister.WorkerJobs(object.GetNamespace()).Get(ownerRef.Name)
		if err != nil {
			klog.V(4).Infof("ignoring orphaned object '%s' of foo '%s'", object.GetSelfLink(), ownerRef.Name)
			return
		}

		h.enqueueWorkerJob(foo)
		return
	}
}
