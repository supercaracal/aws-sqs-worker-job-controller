package controllers

import (
	"fmt"
	"time"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	handlers "github.com/supercaracal/aws-sqs-worker-job-controller/internal/handlers"
	workers "github.com/supercaracal/aws-sqs-worker-job-controller/internal/workers"
	clientset "github.com/supercaracal/aws-sqs-worker-job-controller/pkg/generated/clientset/versioned"
	customscheme "github.com/supercaracal/aws-sqs-worker-job-controller/pkg/generated/clientset/versioned/scheme"
	informers "github.com/supercaracal/aws-sqs-worker-job-controller/pkg/generated/informers/externalversions/awssqsworkerjobcontroller/v1"
	listers "github.com/supercaracal/aws-sqs-worker-job-controller/pkg/generated/listers/awssqsworkerjobcontroller/v1"
)

const controllerAgentName = "aws-sqs-worker-job-controller"

// Controller is
type Controller struct {
	kubeClientSet   kubernetes.Interface
	customClientSet clientset.Interface
	workerJobLister listers.WorkerJobLister
	workerJobSynced cache.InformerSynced
	workQueue       workqueue.RateLimitingInterface
	recorder        record.EventRecorder
}

// NewController is
func NewController(
	kubeClientSet kubernetes.Interface,
	customClientSet clientset.Interface,
	workerJobInformer informers.WorkerJobInfomer) *Controller {

	utilruntime.Must(customscheme.AddToScheme(scheme.Scheme))
	klog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartStructuredLogging(0)
	eventBroadcaster.StartRecordingToSink(
		&typedcorev1.EventSinkImpl{
			Interface: kubeclientset.CoreV1().Events(""),
		},
	)
	recorder := eventBroadcaster.NewRecorder(
		scheme.Scheme,
		corev1.EventSource{
			Component: controllerAgentName,
		},
	)

	wq := workqueue.NewNamedRateLimitingQueue(
		workqueue.DefaultControllerRateLimiter(),
		"WorkerJobs",
	)

	controller := &Controller{
		kubeClientSet:   kubeClientSet,
		customClientSet: customClientSet,
		workerJobLister: workerJobInformer.Lister(),
		workerJobSynced: workerJobInformer.Informer().HasSynced,
		workQueue:       wq,
		recorder:        recorder,
	}

	klog.Info("Setting up event handlers")
	h := handlers.NewInformerHandler(
		controller.WorkerJobLister,
		controller.workQueue,
	)
	workerJobInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    h.OnAdd,
		UpdateFunc: h.OnUpdate,
		DeleteFunc: h.OnDelete,
	})

	return controller
}

// Run is
func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workQueue.ShutDown()

	klog.Info("Starting WorkJob controller")
	klog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.workerJobSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.Info("Starting workers")
	w := workers.NewControllerWorker(
		controller.kubeClientSet,
		controller.customClientSet,
		c.workerJobLister,
		c.workQueue,
		c.record,
	)
	for i := 0; i < threadiness; i++ {
		go wait.Until(w.RunWorker, time.Second, stopCh)
	}

	klog.Info("Started workers")
	<-stopCh
	klog.Info("Shutting down workers")

	return nil
}
