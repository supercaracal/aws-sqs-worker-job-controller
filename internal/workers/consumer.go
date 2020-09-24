package workers

import (
	"context"
	"fmt"
	"strings"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	"github.com/supercaracal/aws-sqs-worker-job-controller/internal/queues"
	customapiv1 "github.com/supercaracal/aws-sqs-worker-job-controller/pkg/apis/awssqsworkerjobcontroller/v1"
	listers "github.com/supercaracal/aws-sqs-worker-job-controller/pkg/generated/listers/awssqsworkerjobcontroller/v1"
)

// Consumer is
type Consumer struct {
	mq     queues.MessageQueue
	cli    kubernetes.Interface
	lister listers.AwsSqsWorkerJobLister
	wq     workqueue.RateLimitingInterface
	rec    record.EventRecorder
}

// NewConsumer is
func NewConsumer(
	region string,
	endpointURL string,
	cli kubernetes.Interface,
	lister listers.AwsSqsWorkerJobLister,
	wq workqueue.RateLimitingInterface,
	rec record.EventRecorder,
) (*Consumer, error) {

	mq, err := queues.NewSQSClient(region, endpointURL)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		mq:     mq,
		cli:    cli,
		lister: lister,
		wq:     wq,
		rec:    rec,
	}, nil
}

// Run is
func (c *Consumer) Run() {
	objs, err := c.lister.List(labels.Everything())
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("Failed to extract custom resource list: %w", err))
		return
	}

	for _, obj := range objs {
		msg, err := c.mq.Dequeue(obj.Spec.QueueURL)
		if err != nil {
			utilruntime.HandleError(fmt.Errorf("Failed to dequeue: %w", err))
			continue
		}
		if msg == "" {
			continue
		}

		enqueueCustomResource(c.wq, obj)

		job, err := createJobResource(obj, msg)
		if err != nil {
			klog.Errorf("Unable to make Job from template in %s/%s: %w", obj.Namespace, obj.Name, err)
			continue
		}

		klog.V(4).Infof("Created Job %s for %s/%s", job.Name, obj.Namespace, obj.Name)
		c.rec.Eventf(cj, v1.EventTypeNormal, "SuccessfulCreate", "Created job %v", job.Name)
	}
}

func enqueueCustomResource(wq workqueue.RateLimitingInterface, obj interface{}) {
	var key string
	var err error

	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}

	wq.Add(key)
}

func createJobResource(obj *customapiv1.AwsSqsWorkerJob, msg string) (*batchv1.Job, error) {
	return c.cli.BatchV1().Jobs(obj.Namespace).Create(
		context.TODO(),
		getJobTemplate(obj, msg),
		metav1.CreateOptions{},
	)
}

func getJobTemplate(obj *customapiv1.AwsSqsWorkerJob, msg string) *batchv1.Job {
	kind := customapiv1.SchemeGroupVersion.WithKind("AwsSqsWorkerJob")

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:            fmt.Sprintf("%s-%d", obj.Name, time.Now().Unix()),
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(obj, kind)},
		},
		Spec: batchv1.JobSpec{
			Parallelism:  1,
			Completions:  1,
			BackoffLimit: 1,
		},
	}

	obj.Spec.Template.DeepCopyInto(&job.Spec.Template)
	job.Spec.Template.Containers[0].Args = strings.Split(msg, " ")

	return job, nil
}
