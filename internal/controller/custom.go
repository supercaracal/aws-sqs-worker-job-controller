package controller

import (
	"fmt"
	"os"
	"time"

	corev1 "k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	kubescheme "k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	batchlisterv1 "k8s.io/client-go/listers/batch/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	handlers "github.com/supercaracal/aws-sqs-worker-job-controller/internal/handler"
	workers "github.com/supercaracal/aws-sqs-worker-job-controller/internal/worker"
	customclient "github.com/supercaracal/aws-sqs-worker-job-controller/pkg/generated/clientset/versioned"
	customscheme "github.com/supercaracal/aws-sqs-worker-job-controller/pkg/generated/clientset/versioned/scheme"
	custominformers "github.com/supercaracal/aws-sqs-worker-job-controller/pkg/generated/informers/externalversions"
	customlisterv1 "github.com/supercaracal/aws-sqs-worker-job-controller/pkg/generated/listers/supercaracal/v1"
)

const (
	informerReSyncDuration = 10 * time.Second
	consumingDuration      = 10 * time.Second
	cleanupDuration        = 10 * time.Second
	resourceName           = "AWSQSWorkerJobs"
)

// CustomController is
type CustomController struct {
	builtin   *builtinTool
	custom    *customTool
	workQueue workqueue.RateLimitingInterface
}

type builtinTool struct {
	client  kubernetes.Interface
	factory kubeinformers.SharedInformerFactory
	job     *jobInfo
}

type customTool struct {
	client   customclient.Interface
	factory  custominformers.SharedInformerFactory
	resource *customResourceInfo
}

type jobInfo struct {
	informer cache.SharedIndexInformer
	lister   batchlisterv1.JobLister
}

type customResourceInfo struct {
	informer cache.SharedIndexInformer
	lister   customlisterv1.AWSSQSWorkerJobLister
}

// NewCustomController is
func NewCustomController(cfg *rest.Config) (*CustomController, error) {
	if err := customscheme.AddToScheme(kubescheme.Scheme); err != nil {
		return nil, err
	}

	builtin, err := buildBuiltinTools(cfg)
	if err != nil {
		return nil, err
	}

	custom, err := buildCustomTools(cfg)
	if err != nil {
		return nil, err
	}

	wq := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), resourceName)
	h := handlers.NewInformerHandler(wq)
	custom.resource.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    h.OnAdd,
		UpdateFunc: h.OnUpdate,
		DeleteFunc: h.OnDelete,
	})

	return &CustomController{builtin: builtin, custom: custom, workQueue: wq}, nil
}

// Run is
func (c *CustomController) Run(stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workQueue.ShutDown()

	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartStructuredLogging(0)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: c.builtin.client.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(kubescheme.Scheme, corev1.EventSource{Component: "controller"})

	c.builtin.factory.Start(stopCh)
	c.custom.factory.Start(stopCh)

	if ok := cache.WaitForCacheSync(stopCh, c.builtin.job.informer.HasSynced, c.custom.resource.informer.HasSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	worker := workers.NewReconciler(
		&workers.ResourceClient{
			Builtin: c.builtin.client,
			Custom:  c.custom.client,
		},
		&workers.ResourceLister{
			Job:            c.builtin.job.lister,
			CustomResource: c.custom.resource.lister,
		},
		c.workQueue,
		recorder,
	)

	if err := worker.WithMessageQueueService(os.Getenv("AWS_REGION"), os.Getenv("AWS_ENDPOINT_URL")); err != nil {
		return err
	}

	go wait.Until(worker.Consume, consumingDuration, stopCh)
	go wait.Until(worker.Clean, cleanupDuration, stopCh)

	klog.V(4).Info("Controller is ready")
	<-stopCh
	klog.V(4).Info("Shutting down controller")

	return nil
}

func buildBuiltinTools(cfg *rest.Config) (*builtinTool, error) {
	cli, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	info := kubeinformers.NewSharedInformerFactory(cli, informerReSyncDuration)
	job := info.Batch().V1().Jobs()
	j := jobInfo{informer: job.Informer(), lister: job.Lister()}

	return &builtinTool{client: cli, factory: info, job: &j}, nil
}

func buildCustomTools(cfg *rest.Config) (*customTool, error) {
	cli, err := customclient.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	info := custominformers.NewSharedInformerFactory(cli, informerReSyncDuration)
	cr := info.Supercaracal().V1().AWSSQSWorkerJobs()
	r := customResourceInfo{informer: cr.Informer(), lister: cr.Lister()}

	return &customTool{client: cli, factory: info, resource: &r}, nil
}
