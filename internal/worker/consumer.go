package worker

import (
	"context"
	"fmt"
	"strings"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/klog/v2"

	queues "github.com/supercaracal/aws-sqs-worker-job-controller/internal/queue"
	customapiv1 "github.com/supercaracal/aws-sqs-worker-job-controller/pkg/apis/supercaracal/v1"
)

var (
	one         int32 = 1
	customGroup       = customapiv1.SchemeGroupVersion.WithKind("AWSSQSWorkerJob")
	creOpts           = metav1.CreateOptions{}
)

// WithMessageQueueService is
func (r *Reconciler) WithMessageQueueService(region string, endpointURL string) (err error) {
	r.messageQueue, err = queues.NewSQSClient(region, endpointURL)
	return
}

// Consume is
func (r *Reconciler) Consume() {
	objs, err := r.lister.CustomResource.List(labels.Everything())
	if err != nil {
		if !kubeerrors.IsNotFound(err) {
			utilruntime.HandleError(err)
		}
		return
	}

	for _, obj := range objs {
		if err := r.dequeueAndCreateJob(obj); err != nil {
			utilruntime.HandleError(err)
		}
	}
}

func (r *Reconciler) dequeueAndCreateJob(obj *customapiv1.AWSSQSWorkerJob) error {
	for {
		msg, err := r.messageQueue.Dequeue(obj.Spec.QueueURL)
		if err != nil {
			return err
		}

		if msg == "" {
			break
		}

		job, err := r.createChildJob(obj, msg)
		if err != nil {
			return fmt.Errorf("Unable to make Job from template in %s/%s: %v", obj.Namespace, obj.Name, err)
		}

		klog.V(4).Infof("Created Job %s for %s/%s", job.Name, obj.Namespace, obj.Name)
		r.recorder.Eventf(obj, corev1.EventTypeNormal, "Successful Create", "Created job %v", job.Name)
	}

	return nil
}

func (r *Reconciler) createChildJob(obj *customapiv1.AWSSQSWorkerJob, msg string) (*batchv1.Job, error) {
	tpl, err := getJobTemplate(obj, msg)
	if err != nil {
		return nil, err
	}

	return r.client.Builtin.BatchV1().Jobs(obj.Namespace).Create(context.TODO(), tpl, creOpts)
}

func getJobTemplate(obj *customapiv1.AWSSQSWorkerJob, msg string) (*batchv1.Job, error) {
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:            fmt.Sprintf("%s-%d", obj.Name, time.Now().UnixMicro()),
			Namespace:       obj.Namespace,
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(obj, customGroup)},
		},
		Spec: batchv1.JobSpec{
			Parallelism:  &one,
			Completions:  &one,
			BackoffLimit: &one,
		},
	}

	obj.Spec.Template.DeepCopyInto(&job.Spec.Template)
	if len(job.Spec.Template.Spec.Containers) == 0 {
		return nil, fmt.Errorf("failed to copy custom resource data, make sure the OpenAPI schema in your CRD manifest")
	}

	job.Spec.Template.Spec.Containers[0].Args = strings.Split(msg, " ")
	job.Spec.Template.Spec.RestartPolicy = "Never"

	return job, nil
}
