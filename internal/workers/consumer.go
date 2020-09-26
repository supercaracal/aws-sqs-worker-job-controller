package workers

import (
	"context"
	"fmt"
	"strings"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
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
	ns     string
}

// NewConsumer is
func NewConsumer(
	region string,
	endpointURL string,
	selfNamespace string,
	cli kubernetes.Interface,
	lister listers.AwsSqsWorkerJobLister,
	wq workqueue.RateLimitingInterface,
	rec record.EventRecorder,
) (*Consumer, error) {

	if selfNamespace == "" {
		return nil, fmt.Errorf("SELF_NAMESPACE env var requireed")
	}

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
		ns:     selfNamespace,
	}, nil
}

// Run is
func (c *Consumer) Run() {
	objs, err := c.lister.AwsSqsWorkerJobs(c.ns).List(labels.Everything())
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

		c.enqueueCustomResource(obj)

		job, err := c.createJobResource(obj, msg)
		if err != nil {
			klog.Errorf("Unable to make Job from template in %s/%s: %v", obj.Namespace, obj.Name, err)
			continue
		}

		klog.V(4).Infof("Created Job %s for %s/%s", job.Name, obj.Namespace, obj.Name)
		c.rec.Eventf(obj, corev1.EventTypeNormal, "SuccessfulCreate", "Created job %v", job.Name)
	}
}

func (c *Consumer) enqueueCustomResource(obj interface{}) {
	var key string
	var err error

	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}

	c.wq.Add(key)
}

func (c *Consumer) createJobResource(obj *customapiv1.AwsSqsWorkerJob, msg string) (*batchv1.Job, error) {
	tpl, err := getJobTemplate(obj, msg)
	if err != nil {
		return nil, err
	}

	return c.cli.BatchV1().Jobs(obj.Namespace).Create(context.TODO(), tpl, metav1.CreateOptions{})
}

func getJobTemplate(obj *customapiv1.AwsSqsWorkerJob, msg string) (*batchv1.Job, error) {
	var one int32 = 1
	kind := customapiv1.SchemeGroupVersion.WithKind("AwsSqsWorkerJob")

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:            fmt.Sprintf("%s-%d", obj.Name, time.Now().Unix()),
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(obj, kind)},
		},
		Spec: batchv1.JobSpec{
			Parallelism:  &one,
			Completions:  &one,
			BackoffLimit: &one,
		},
	}

	obj.Spec.Template.DeepCopyInto(&job.Spec.Template)
	if len(job.Spec.Template.Spec.Containers) == 0 {
		return nil, fmt.Errorf("failed to copy custom resource data, make sure the OpenAPI schema in your manifest")
	}

	job.Spec.Template.Spec.Containers[0].Args = strings.Split(msg, " ")
	job.Spec.Template.Spec.RestartPolicy = "Never"

	return job, nil
}
