package controllers

import (
	"fmt"
	"os"
	"time"

	corev1 "k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	batchinformers "k8s.io/client-go/informers/batch/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	batchlisters "k8s.io/client-go/listers/batch/v1"
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

// CustomController is
type CustomController struct {
	kubeClientSet        kubernetes.Interface
	customClientSet      clientset.Interface
	jobLister            batchlisters.JobLister
	jobSynced            cache.InformerSynced
	customResourceLister listers.AwsSqsWorkerJobLister
	customInformerSynced cache.InformerSynced
	workQueue            workqueue.RateLimitingInterface
	recorder             record.EventRecorder
}

// NewCustomController is
func NewCustomController(
	kubeClientSet kubernetes.Interface,
	customClientSet clientset.Interface,
	jobInformer batchinformers.JobInformer,
	customInformer informers.AwsSqsWorkerJobInformer,
) *CustomController {

	utilruntime.Must(customscheme.AddToScheme(scheme.Scheme))
	klog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartStructuredLogging(0)
	eventBroadcaster.StartRecordingToSink(
		&typedcorev1.EventSinkImpl{
			Interface: kubeClientSet.CoreV1().Events(""),
		},
	)
	recorder := eventBroadcaster.NewRecorder(
		scheme.Scheme,
		corev1.EventSource{
			Component: "aws-sqs-worker-job-controller",
		},
	)

	wq := workqueue.NewNamedRateLimitingQueue(
		workqueue.DefaultControllerRateLimiter(),
		"AwsSqsWorkerJobs",
	)

	controller := &CustomController{
		kubeClientSet:        kubeClientSet,
		customClientSet:      customClientSet,
		jobLister:            jobInformer.Lister(),
		jobSynced:            jobInformer.Informer().HasSynced,
		customResourceLister: customInformer.Lister(),
		customInformerSynced: customInformer.Informer().HasSynced,
		workQueue:            wq,
		recorder:             recorder,
	}

	klog.Info("Setting up event handlers")
	h := handlers.NewInformerHandler()
	customInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    h.OnAdd,
		UpdateFunc: h.OnUpdate,
		DeleteFunc: h.OnDelete,
	})

	return controller
}

// Run is
func (c *CustomController) Run(stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workQueue.ShutDown()

	klog.Info("Starting AwsSqsWorkerJob controller")
	klog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.jobSynced, c.customInformerSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.Info("Starting workers")
	rw := workers.NewReconciler(
		c.kubeClientSet,
		c.customClientSet,
		c.jobLister,
		c.customResourceLister,
		c.workQueue,
		c.recorder,
	)
	cw, err := workers.NewConsumer(
		os.Getenv("AWS_REGION"),
		os.Getenv("AWS_ENDPOINT_URL"),
		c.kubeClientSet,
		c.customResourceLister,
		c.workQueue,
		c.recorder,
	)
	if err != nil {
		return fmt.Errorf("failed to initialize consumer worker: %w", err)
	}

	go wait.Until(rw.Run, 10*time.Second, stopCh)
	go wait.Until(cw.Run, 10*time.Second, stopCh)

	klog.Info("Started workers")
	<-stopCh
	klog.Info("Shutting down workers")

	return nil
}
